// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package policyfilter

import (
	"fmt"
	"sync"

	slimv1 "github.com/cilium/cilium/pkg/k8s/slim/k8s/apis/meta/v1"
	"github.com/cilium/tetragon/pkg/cgroups/fsscan"
	"github.com/cilium/tetragon/pkg/labels"
	"github.com/cilium/tetragon/pkg/logger"
	"github.com/cilium/tetragon/pkg/metrics/policyfiltermetrics"
	"github.com/cilium/tetragon/pkg/podhooks"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
)

// Policy filter is a mechanism for restricting  tracing policies on a subset
// of pods running in the node. Policies are identified by their policyID and
// the pod processes are identified the cgroup ids of their containers.
//
// The pods that match a given policy are selected based on:
//   (1) Namespaces
//   (2) Label filters
//
// This package maintains the 'policy_filter_maps' bpf map. Bpf checks this map
// to decide whether a policy is applied or not. The map is a hash-of-hashes:
//
//   policy_id -> [ cgroup_id -> u8 ]
//
// If entry policy_id -> cgroup_id exists, then policy is to be applied. (u8 value is ignored for
// now.)
//
// This package provides functions that can be used to update the bpf map. The map
// needs to be updated in the following conditions:
//
//  (A) Policy changes: when a new policy is added (or deleted), we need to add cgroup ids of
//  matching pods. See {Add,Del}Policy.
//
//  (B) Pod containers changes: when new containers are added (or deleted): we need to add the cgroup
//  ids of matching policies. See AddPodContainer, DelPodContainer, DelPod.
//
//  (C) Pod labels change: need to rescan policies because the result of pod label filters might have
//  changed.
//
// Todo:
//  - use a goroutine and a queue
//  (https://github.com/kubernetes/client-go/blob/master/examples/workqueue/main.go) instead locks
//  for serialization

func init() {
	podhooks.RegisterCallbacksAtInit(podhooks.Callbacks{
		PodCallbacks: func(podInformer cache.SharedIndexInformer) {
			// register pod handlers for policyfilters
			if pfState, err := GetState(); err == nil {
				logger.GetLogger().Info("registering policyfilter pod handlers")
				pfState.RegisterPodHandlers(podInformer)
			}
		},
	})
}

const (
	// polMapSize is the number of entries for the (inner) policy map. It
	// should be large enough to accommodate the number of containers
	// running in a system.
	polMapSize = 32768
)

type PolicyID uint32
type PodID uuid.UUID
type CgroupID uint64

const (
	// we reserve 0 as a special value to indicate no filtering
	NoFilterPolicyID         = 0
	NoFilterID               = PolicyID(NoFilterPolicyID)
	FirstValidFilterPolicyID = NoFilterPolicyID + 1
)

func (i PodID) String() string {
	var x uuid.UUID = uuid.UUID(i)
	return x.String()
}

type containerInfo struct {
	id   string   // container id
	cgID CgroupID // cgroup id
}

// podInfo contains the necessary information for each pod
type podInfo struct {
	id         PodID
	namespace  string
	labels     labels.Labels
	containers []containerInfo

	// cache of matched policies
	matchedPolicies []PolicyID
}

func (pod *podInfo) cgroupIDs() []CgroupID {
	ret := make([]CgroupID, 0, len(pod.containers))
	for i := range pod.containers {
		ret = append(ret, pod.containers[i].cgID)
	}
	return ret
}

// delete containers from a pod based on their id, and return them
// NB: in most cases there will be a single container, but we do not reject users adding a container
// with the same id and different cgroup, so we return a list to cover all cases.
func (pod *podInfo) delContainers(id string) []containerInfo {
	var ret []containerInfo
	for i := 0; i < len(pod.containers); i++ {
		c := pod.containers[i]
		if c.id == id {
			ret = append(ret, c)
			pod.containers = append(pod.containers[:i], pod.containers[i+1:]...)
			i--
		}
	}
	return ret
}

func (pod *podInfo) delCachedPolicy(polID PolicyID) {
	for i := 0; i < len(pod.matchedPolicies); i++ {
		if pod.matchedPolicies[i] == polID {
			pod.matchedPolicies = append(pod.matchedPolicies[:i], pod.matchedPolicies[i+1:]...)
		}
	}
}

func (pod *podInfo) addCachedPolicy(polID PolicyID) {
	for i := 0; i < len(pod.matchedPolicies); i++ {
		if pod.matchedPolicies[i] == polID {
			return
		}
	}
	pod.matchedPolicies = append(pod.matchedPolicies, polID)
}

func (pod *podInfo) hasPolicy(polID PolicyID) bool {
	for i := 0; i < len(pod.matchedPolicies); i++ {
		if pod.matchedPolicies[i] == polID {
			return true
		}
	}
	return false
}

// containerExists checks returns true if a container exists in the pod
func (m *state) containerExists(pod *podInfo, containerID string, cgIDp *CgroupID) bool {
	for i := range pod.containers {
		container := &pod.containers[i]
		if container.id != containerID {
			continue
		}

		// we found a matching container id, if no cgroup id is given, return
		if cgIDp == nil {
			return true
		}

		// otherwise, also check that the cgroup id matches
		if container.id == containerID && container.cgID == *cgIDp {
			// container was already handled, return
			return true
		}

		// issue  a warning and continue
		// NB: if this happens, we might end up with multiple cgroup
		// ids for the same container. Since we have no way of knowing
		// which one is the correct, we keep both.
		m.log.WithFields(logrus.Fields{
			"pod-id":        pod.id,
			"container-id":  containerID,
			"old-cgroup-id": container.cgID,
			"new-cgroup-id": *cgIDp,
		}).Warnf("AddPodContainer: conflicting cgroup ids")
	}

	return false
}

type policy struct {
	id PolicyID

	// if namespace is "", policy applies to all namespaces
	namespace string

	podSelector labels.Selector

	// polMap is the (inner) policy map for this policy
	polMap polMap
}

func (pol *policy) podMatches(podNs string, podLabels labels.Labels) bool {
	if pol.namespace != "" && podNs != pol.namespace {
		return false
	}

	return pol.podSelector.Match(podLabels)
}

func (pol *policy) podInfoMatches(pod *podInfo) bool {
	return pol.podMatches(pod.namespace, pod.labels)
}

// State holds the necessary state for policyfilter
type state struct {
	log logrus.FieldLogger

	// mutex serializes access to the internal structures, as well as operations.
	mu       sync.Mutex
	policies []policy
	pods     []podInfo

	// polify filters (outer) map handle
	pfMap PfMap

	cgidFinder cgidFinder
}

// New creates a new State of the policy filter code. Callers should call Close() to release
// allocated resources (namely the bpf map).
//
//revive:disable:unexported-return
func New() (*state, error) {
	log := logger.GetLogger().WithField("subsystem", "policy-filter")
	return newState(
		log,
		&cgfsFinder{fsscan.New(), log},
	)
}

func newState(
	log logrus.FieldLogger,
	cgidFinder cgidFinder,
) (*state, error) {
	var err error
	ret := &state{
		log:        log,
		cgidFinder: cgidFinder,
	}

	ret.pfMap, err = newPfMap()
	if err != nil {
		return nil, err
	}

	return ret, nil
}

//revive:enable:unexported-return

func (m *state) updatePodHandler(pod *v1.Pod) error {
	containerIDs := podContainersIDs(pod)
	podID, err := uuid.Parse(string(pod.UID))
	if err != nil {
		m.log.WithError(err).WithField("pod-id", pod.UID).Warn("policyfilter, pod handler: failed to parse pod id")
		return err
	}

	namespace := pod.Namespace
	err = m.UpdatePod(PodID(podID), namespace, pod.Labels, containerIDs)
	if err != nil {
		m.log.WithError(err).WithFields(logrus.Fields{
			"pod-id":        podID,
			"container-ids": containerIDs,
			"namespace":     namespace,
		}).Warn("policyfilter, add pod-handler: AddPodContainer failed")
		return err
	}

	return nil
}

func (m *state) getPodEventHandlers() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				logger.GetLogger().Warn("policyfilter, add-pod handler: unexpected object type: %T", pod)
				return
			}
			err := m.updatePodHandler(pod)
			policyfiltermetrics.OpInc("pod-handlers", "add", err)
		},
		UpdateFunc: func(_, newObj interface{}) {
			pod, ok := newObj.(*v1.Pod)
			if !ok {
				logger.GetLogger().Warn("policyfilter, update-pod: unexpected object type(s): new:%T", pod)
				return
			}
			err := m.updatePodHandler(pod)
			policyfiltermetrics.OpInc("pod-handlers", "update", err)
		},
		DeleteFunc: func(obj interface{}) {
			// Remove all containers for this pod
			pod, ok := obj.(*v1.Pod)
			if !ok {
				logger.GetLogger().Warn("policyfilter, delete-pod handler: unexpected object type: %T", pod)
				return
			}
			podID, err := uuid.Parse(string(pod.UID))
			if err != nil {
				logger.GetLogger().WithField("pod-id", pod.UID).WithError(err).Warn("policyfilter, delete-pod: failed to parse id")
				return
			}

			namespace := pod.Namespace
			err = m.DelPod(PodID(podID))
			if err != nil {
				logger.GetLogger().WithError(err).WithFields(logrus.Fields{
					"pod-id":    podID,
					"namespace": namespace,
				}).Warn("policyfilter, delete-pod handler: DelPod failed")
			}
			policyfiltermetrics.OpInc("pod-handlers", "delete", err)
		},
	}
}

func (m *state) RegisterPodHandlers(podInformer cache.SharedIndexInformer) {
	podInformer.AddEventHandler(m.getPodEventHandlers())
}

// Close releases resources allocated by the Manager. Specifically, we close and unpin the policy filter map.
func (m *state) Close() error {
	return m.pfMap.release()
}

func (m *state) findPolicy(id PolicyID) *policy {
	for i := range m.policies {
		if m.policies[i].id == id {
			return &m.policies[i]
		}
	}
	return nil
}

// delPolicy removes a policy and returns it, or returns nil if policy is not found
func (m *state) delPolicy(id PolicyID) *policy {
	for i, pol := range m.policies {
		if pol.id == id {
			m.policies = append(m.policies[:i], m.policies[i+1:]...)
			return &pol
		}
	}
	return nil
}

// find the pod with the given id
func (m *state) findPod(id PodID) *podInfo {
	for i := range m.pods {
		if m.pods[i].id == id {
			return &m.pods[i]
		}
	}
	return nil
}

func (m *state) delPod(id PodID) *podInfo {
	for i, pod := range m.pods {
		if pod.id == id {
			m.pods = append(m.pods[:i], m.pods[i+1:]...)
			return &pod
		}
	}
	return nil
}

// AddPolicy adds a policy
func (m *state) AddPolicy(polID PolicyID, namespace string, podLabelSelector *slimv1.LabelSelector) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if p := m.findPolicy(polID); p != nil {
		return fmt.Errorf("policy with id %d already exists: not adding new one", polID)
	}

	podSelector, err := labels.SelectorFromLabelSelector(podLabelSelector)
	if err != nil {
		return err
	}
	policy := policy{
		id:          polID,
		namespace:   namespace,
		podSelector: podSelector,
	}

	cgroupIDs := make([]CgroupID, 0)
	// scan pods to find the ones that match this policy to set initial state for policy
	matchedPods := make([]*podInfo, 0, len(m.pods))
	for i := range m.pods {
		pod := &m.pods[i]
		if !policy.podInfoMatches(pod) {
			continue
		}
		for cIdx := range pod.containers {
			cgroupIDs = append(cgroupIDs, pod.containers[cIdx].cgID)
		}
		matchedPods = append(matchedPods, pod)
		pod.addCachedPolicy(policy.id)
	}

	// update state for policy
	policy.polMap, err = m.pfMap.newPolicyMap(polID, cgroupIDs)
	if err != nil {
		for _, pod := range matchedPods {
			pod.delCachedPolicy(policy.id)
		}
		return fmt.Errorf("adding policy data to map failed: %w", err)
	}

	m.policies = append(m.policies, policy)

	return nil
}

// DelPolicy will destroly all information for the provided policy
func (m *state) DelPolicy(polID PolicyID) error {

	if polID == NoFilterPolicyID {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	policy := m.delPolicy(polID)
	if policy != nil {
		policy.polMap.Close()
	} else {
		m.log.WithField("policy-id", polID).Warn("DelPolicy: policy internal map not found")
	}

	if err := m.pfMap.Delete(polID); err != nil {
		m.log.WithField("policy-id", polID).Warn("DelPolicy: failed to remove policy from external map")
	}

	for i := range m.pods {
		pod := &m.pods[i]
		pod.delCachedPolicy(policy.id)
	}

	return nil
}

func cgIDPointerStr(p *CgroupID) string {
	if p == nil {
		return "(unknown)"
	}
	return fmt.Sprintf("%d", *p)
}

// addPodContainers adds a list of containers (ids) to a pod.
// It will update the state for all containers that do not exist.
// It takes an optional argument of a list of cgroup ids (one per container). If this list is empty,
// the function will try to figure out the cgroup id on its own.
// Finally, it will scan over all the matching policies for the pod and update the policy maps.
func (m *state) addPodContainers(pod *podInfo, containerIDs []string, cgroupIDs []CgroupID) {
	// Find the containers that do not exist in our state, and for those find the cgroup id if
	// one does not exist.
	cinfo := make([]containerInfo, 0, len(containerIDs))
	cgIDs := make([]CgroupID, 0, len(containerIDs))
	for i, contID := range containerIDs {
		var cgIDptr *CgroupID
		if len(cgroupIDs) > i {
			cgIDptr = &cgroupIDs[i]
		}

		if m.containerExists(pod, contID, cgIDptr) {
			m.debugLogWithCallers(4).WithFields(logrus.Fields{
				"pod-id":       pod.id,
				"namespace":    pod.namespace,
				"container-id": contID,
				"cgroup-id":    cgIDPointerStr(cgIDptr),
			}).Info("addPodContainers: container exists, skipping")
			continue
		}

		if cgIDptr == nil {
			cgid, err := m.cgidFinder.findCgroupID(pod.id, contID)
			if err != nil {
				// error: skip this container id
				m.log.WithError(err).WithFields(logrus.Fields{
					"pod-id":       pod.id,
					"container-id": contID,
				}).Warn("failed to find cgroup id. Skipping container.")
				continue
			}
			cgIDptr = &cgid
		}

		cinfo = append(cinfo, containerInfo{contID, *cgIDptr})
		cgIDs = append(cgIDs, *cgIDptr)
	}

	if len(cinfo) == 0 {
		m.debugLogWithCallers(4).WithFields(logrus.Fields{
			"pod-id":        pod.id,
			"namespace":     pod.namespace,
			"container-ids": containerIDs,
		}).Info("addPodContainers: nothing to do, returning")
		return
	}

	// update containers
	pod.containers = append(pod.containers, cinfo...)
	m.debugLogWithCallers(4).WithFields(logrus.Fields{
		"pod-id":          pod.id,
		"namespace":       pod.namespace,
		"containers-info": cinfo,
	}).Info("addPodContainers: container(s) added")

	// update matching policy maps
	for _, policyID := range pod.matchedPolicies {
		pol := m.findPolicy(policyID)
		if pol == nil {
			m.log.WithFields(logrus.Fields{
				"policy-id":  policyID,
				"pod-id":     pod.id,
				"cgroup-ids": cgroupIDs,
			}).Warn("addPodContainers: unknown policy id found in pod. This should not happen, ignoring.")
			continue
		}

		if err := pol.polMap.addCgroupIDs(cgIDs); err != nil {
			m.log.WithError(err).WithFields(logrus.Fields{
				"policy-id":  pol.id,
				"pod-id":     pod.id,
				"cgroup-ids": cgroupIDs,
			}).Warn("failed to update policy map")
		}
	}
}

func (m *state) addNewPod(podID PodID, namespace string, podLabels labels.Labels) *podInfo {
	m.pods = append(m.pods, podInfo{
		id:         podID,
		namespace:  namespace,
		labels:     podLabels,
		containers: nil,
	})
	pod := &m.pods[len(m.pods)-1]
	for i := range m.policies {
		pol := &m.policies[i]
		if pol.podInfoMatches(pod) {
			pod.addCachedPolicy(pol.id)
		}
	}
	return pod
}

// AddPodContainer informs policyfilter about a new container in a pod.
// if the cgroup id of the container is known, cgID is not nil and it contains its value.
//
// The pod might or might not have been encountered before.
func (m *state) AddPodContainer(podID PodID, namespace string, podLabels labels.Labels, containerID string, cgID CgroupID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pod := m.findPod(podID)
	if pod == nil {
		pod = m.addNewPod(podID, namespace, podLabels)
		m.debugLogWithCallers(4).WithFields(logrus.Fields{
			"pod-id":       podID,
			"namespace":    namespace,
			"container-id": containerID,
			"cgroup-id":    cgID,
		}).Info("AddPodContainer: added pod")
	} else if pod.namespace != namespace {
		// sanity check: old and new namespace should match
		return fmt.Errorf("conflicting namespaces for pod with id %s: old='%s' vs new='%s'", podID, pod.namespace, namespace)
	}

	m.addPodContainers(pod, []string{containerID}, []CgroupID{cgID})
	return nil
}

// delPodCgroupIDsFromPolicyMaps will delete cgorup entries for containers belonging to pod on all
// policy maps.
func (m *state) delPodCgroupIDsFromPolicyMaps(pod *podInfo, containers []containerInfo) {

	if len(containers) == 0 {
		return
	}

	cgroupIDs := make([]CgroupID, 0, len(containers))
	for i := range containers {
		cgroupIDs = append(cgroupIDs, containers[i].cgID)
	}

	// check what policies match the pod, and delete the cgroup ids
	for _, policyID := range pod.matchedPolicies {
		pol := m.findPolicy(policyID)
		if pol == nil {
			m.log.WithFields(logrus.Fields{
				"policy-id":  policyID,
				"pod-id":     pod.id,
				"cgroup-ids": cgroupIDs,
			}).Warn("delPodCgroupIDsFromPolicyMaps: unknown policy id found in pod. This should not happen, ignoring.")
			continue
		}

		if err := pol.polMap.delCgroupIDs(cgroupIDs); err != nil {
			// NB: depending on the error, we might want to schedule some retries here
			m.log.WithError(err).WithFields(logrus.Fields{
				"policy-id":  pol.id,
				"pod-id":     pod.id,
				"cgroup-ids": cgroupIDs,
			}).Warn("delPodCgroupIDsFromPolicyMaps: failed to delete cgroup ids from policy map")
		}
	}
}

// DelPodContainer informs policyfilter that a container was deleted from a pod
func (m *state) DelPodContainer(podID PodID, containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pod := m.findPod(podID)
	if pod == nil {
		m.Debugf("DelPodContainer: pod-id %s not found", podID)
		return nil
	}

	containers := pod.delContainers(containerID)
	if len(containers) != 1 {
		m.Debugf("DelPodContainer: pod-id=%s container-id=%s had %d containers: %v", podID, containerID, len(containers), containers)
	}
	m.delPodCgroupIDsFromPolicyMaps(pod, containers)
	return nil
}

// DelPod informs policyfilter that a pod has been deleted
func (m *state) DelPod(podID PodID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pod := m.delPod(podID)
	if pod == nil {
		m.Debug("DelPod: pod-id %s not found", podID)
		return nil
	}
	m.delPodCgroupIDsFromPolicyMaps(pod, pod.containers)
	return nil
}

// policiesDiffRes is the result returned by policiesDiff
type policiesDiffRes struct {
	// addedPolicies is a slice of the policies that were added
	addedPolicies []*policy
	// deletedPolicies is a slice of the policies that were removed
	deletedPolicies []*policy
	// newMatchedPolicies is a slice of the policy ids that now match the pod
	newMatchedPolicies []PolicyID
}

func (m *state) policiesDiff(pod *podInfo, newLabels labels.Labels) *policiesDiffRes {
	addedPolicies := []*policy{}
	deletedPolicies := []*policy{}
	newMatchedPolicies := []PolicyID{}
	for i := range m.policies {
		pol := &m.policies[i]
		podHasPolicy := pod.hasPolicy(pol.id)
		if pol.podMatches(pod.namespace, newLabels) {
			newMatchedPolicies = append(newMatchedPolicies, pol.id)
			if !podHasPolicy {
				// policy matches, but pod does not have it in its matched policies.
				addedPolicies = append(addedPolicies, pol)
			}
		} else if podHasPolicy {
			// policy does not match, but pod already has it as matched
			deletedPolicies = append(deletedPolicies, pol)
		}
	}

	return &policiesDiffRes{
		addedPolicies:      addedPolicies,
		deletedPolicies:    deletedPolicies,
		newMatchedPolicies: newMatchedPolicies,
	}
}

// applyPodPolicyDiff applies the changes of the policies that match a pod to the state by updating:
// - the policy maps
//   - adding cgroup ids for policies that match now, but did not match before
//   - removing cgroup ids for policies that did not match, but did match before
//
// - pod.matchedPolicies with a slice of the new policy ids
func (m *state) applyPodPolicyDiff(pod *podInfo, polDiff *policiesDiffRes) {
	// no changes, just return
	if len(polDiff.addedPolicies) == 0 && len(polDiff.deletedPolicies) == 0 {
		return
	}

	cgroupIDs := pod.cgroupIDs()
	for _, addPol := range polDiff.addedPolicies {
		if err := addPol.polMap.addCgroupIDs(cgroupIDs); err != nil {
			m.log.WithError(err).WithFields(logrus.Fields{
				"policy-id":  addPol.id,
				"pod-id":     pod.id,
				"cgroup-ids": cgroupIDs,
				"reason":     "labels change caused policy to match",
			}).Warn("failed to update policy map")
		}
	}

	for _, delPol := range polDiff.deletedPolicies {
		if err := delPol.polMap.delCgroupIDs(cgroupIDs); err != nil {
			m.log.WithError(err).WithFields(logrus.Fields{
				"policy-id":  delPol.id,
				"pod-id":     pod.id,
				"cgroup-ids": cgroupIDs,
				"reason":     "labels change caused policy to unmatch",
			}).Warn("failed to update policy map")
		}
	}
	pod.matchedPolicies = polDiff.newMatchedPolicies
}

func (pod *podInfo) containerDiff(newContainerIDs []string) ([]string, []string) {

	// maintain a hash of new ids. The values indicate whether the id was seen in existing ids
	// or not
	newIDs := make(map[string]bool)
	for _, cid := range newContainerIDs {
		newIDs[cid] = false
	}

	addContIDs := []string{}
	delContIDs := []string{}
	for _, containerInfo := range pod.containers {
		cid := containerInfo.id
		if _, exists := newIDs[cid]; exists {
			newIDs[cid] = true
		} else {
			delContIDs = append(delContIDs, cid)
		}
	}

	for cid, seen := range newIDs {
		if !seen {
			addContIDs = append(addContIDs, cid)
		}
	}

	return addContIDs, delContIDs
}

// UpdatePod updates the pod state for a pod
// containerIDs contains all the running container ids for the given pod.
// This function will:
//   - remove the containers that are not part of the containerIDs list
//   - add the ones that do not exist in the current state
//
// It is intended to be used from k8s watchers (where no cgroup information is available)
func (m *state) UpdatePod(podID PodID, namespace string, podLabels labels.Labels, containerIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dlog := m.debugLogWithCallers(4).WithFields(logrus.Fields{
		"pod-id":        podID,
		"namespace":     namespace,
		"container-ids": containerIDs,
	})

	pod := m.findPod(podID)
	if pod == nil {
		pod = m.addNewPod(podID, namespace, podLabels)
		dlog.Info("UpdatePod: added pod")
	} else if pod.namespace != namespace {
		// sanity check: old and new namespace should match
		return fmt.Errorf("conflicting namespaces for pod with id %s: old='%s' vs new='%s'", podID, pod.namespace, namespace)
	}

	// labels changed: check if there are policies ads that:
	// - did not match before, but they match now (addPols)
	// - did match before, but they do not match now (delPols))
	// and update state accordingly
	if pod.labels.Cmp(podLabels) {
		polDiff := m.policiesDiff(pod, podLabels)
		m.debugLogWithCallers(1).WithFields(logrus.Fields{
			"pod-id":         pod.id,
			"pod-old-labels": pod.labels,
			"pod-new-labels": podLabels,
			"policy-diff":    fmt.Sprintf("%+v", polDiff),
		}).Info("UpdatePod: pod labels changed")
		m.applyPodPolicyDiff(pod, polDiff)
		pod.labels = podLabels
	}

	// containers changed: check if there are new or deleted containers, and update the policy
	// map
	addIDs, delIDs := pod.containerDiff(containerIDs)
	for _, cid := range delIDs {
		containers := pod.delContainers(cid)
		if len(containers) != 1 {
			m.Debugf("UpdatePod: pod-id=%s container-id=%s had %d containers: %v", podID, cid, len(containers), containers)
		}
		m.delPodCgroupIDsFromPolicyMaps(pod, containers)
	}

	m.addPodContainers(pod, addIDs, nil)
	return nil
}
