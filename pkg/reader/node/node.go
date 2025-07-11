// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package node

import (
	"os"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/pkg/logger"
	"github.com/cilium/tetragon/pkg/logger/logfields"
	"github.com/cilium/tetragon/pkg/option"
)

const (
	hubbleNodeNameEnvVar = "HUBBLE_NODE_NAME"
	nodeNameEnvVar       = "NODE_NAME"
)

var (
	nodeName       string
	exportNodeName string
	nodeLabels     = map[string]string{}
)

func init() {
	SetExportNodeName()
	SetNodeName()
}

// SetExportNodeName initializes the exportNodeName variable. It's defined separately from
// init() so that it can be called from unit tests.
func SetExportNodeName() {
	var err error
	if exportNodeName = os.Getenv(hubbleNodeNameEnvVar); exportNodeName != "" {
		return
	}
	if exportNodeName = os.Getenv(nodeNameEnvVar); exportNodeName != "" {
		return
	}
	exportNodeName, err = os.Hostname()
	if err != nil {
		logger.GetLogger().Warn("failed to retrieve hostname", logfields.Error, err)
	}
}

// SetNodeName initializes the nodeName variable. It's defined separately from
// init() so that it can be called from unit tests.
func SetNodeName() {
	var err error
	if nodeName = os.Getenv(nodeNameEnvVar); nodeName != "" {
		return
	}
	if nodeName = os.Getenv(hubbleNodeNameEnvVar); nodeName != "" {
		return
	}
	nodeName, err = os.Hostname()
	if err != nil {
		logger.GetLogger().Warn("failed to retrieve hostname", logfields.Error, err)
	}
}

func SetNodeLabels(labels map[string]string) {
	nodeLabels = labels
}

// GetNodeNameForExport returns node name string for JSON export. It uses the HUBBLE_NODE_NAME
// env variable by default, and falls back to NODE_NAME if the former is missing. If both
// are missing, it will use the host name reported by the kernel
func GetNodeNameForExport() string {
	return exportNodeName
}

// GetNodeName returns node name string for the given node in Kubernetes. It uses the NODE_NAME
// env variable by default, and falls back to HUBBLE_NODE_NAME if the former is missing. If both
// are missing, it will use the host name reported by the kernel. This value is used when watching for
// pods running on the node in Kubernetes.
//
// NOTE: This is different from the Export equivalent for cases where nodes in kubernetes are named different
// from the desired node name in the JSON export.
func GetNodeName() string {
	return nodeName
}

// SetCommonFields set fields that are common in all the events.
func SetCommonFields(ev *tetragon.GetEventsResponse) {
	ev.NodeName = exportNodeName
	ev.ClusterName = option.Config.ClusterName
	ev.NodeLabels = nodeLabels
}
