---
title: "Metrics"
description: >
  This reference is autogenerated from the Tetragon Prometheus metrics registry.
weight: 4
---
{{< comment >}}
This page is autogenerated via `make metrics-doc` please do not edit directly.
{{< /comment >}}
## Tetragon Health Metrics

### `tetragon_bpf_missed_events_total`

Number of Tetragon perf events that are failed to be sent from the kernel.

| label | values |
| ----- | ------ |
| `error` | `E2BIG, EBUSY, EINVAL, ENOENT, ENOSPC, unknown` |
| `msg_op` | `13, 14, 15, 16, 23, 24, 25, 26, 27, 5, 7` |

### `tetragon_build_info`

Build information about tetragon

| label | values |
| ----- | ------ |
| `commit` | `931b70f2c9878ba985ba6b589827bea17da6ec33` |
| `go_version` | `go1.24.5` |
| `modified` | `false` |
| `time ` | `2022-05-13T15:54:45Z` |
| `version` | `v1.2.0` |

### `tetragon_cri_cgidmap_resolutions_errors_total`

number of cgroup id map (cgidmap) CRI resolutions that failed

### `tetragon_cri_cgidmap_resolutions_total`

number of total cgroup id map (cgidmap) CRI resolutions

### `tetragon_data_cache_capacity`

The capacity of the data cache.

### `tetragon_data_cache_evictions_total`

Number of data cache LRU evictions.

### `tetragon_data_cache_misses_total`

Number of data cache misses.

| label | values |
| ----- | ------ |
| `operation` | `get, remove` |

### `tetragon_data_cache_size`

The size of the data cache

### `tetragon_data_event_size`

The size of received data events.

| label | values |
| ----- | ------ |
| `op   ` | `bad, ok` |

### `tetragon_data_events_total`

The number of data events by type. For internal use only.

| label | values |
| ----- | ------ |
| `event` | `Added, Appended, Bad, Matched, NotMatched, Received` |

### `tetragon_enforcer_missed_notifications_total`

The number of missed notifications by the enforcer.

| label | values |
| ----- | ------ |
| `info ` | `syscall` |
| `policy` | `policy-name` |
| `reason` | `reason` |

### `tetragon_errors_total`

The total number of Tetragon errors. For internal use only.

| label | values |
| ----- | ------ |
| `type ` | `event_finalize_process_info_failed, process_metadata_username_failed, process_metadata_username_ignored_not_in_host_namespaces, process_pid_tid_mismatch` |

### `tetragon_event_cache_entries`

The number of entries in the event cache.

### `tetragon_event_cache_errors_total`

The total of errors encountered while fetching process exec information from the cache.

| label | values |
| ----- | ------ |
| `error` | `nil_process_pid` |
| `event_type` | `PROCESS_EXEC, PROCESS_EXIT, PROCESS_KPROBE, PROCESS_LOADER, PROCESS_LSM, PROCESS_THROTTLE, PROCESS_TRACEPOINT, PROCESS_UPROBE, RATE_LIMIT_INFO` |

### `tetragon_event_cache_fetch_failures_total`

Number of failed fetches from the event cache. These won't be retried as they already exceeded the limit.

| label | values |
| ----- | ------ |
| `entry_type` | `ancestors_info, parent_info, pod_info, process_info` |
| `event_type` | `PROCESS_EXEC, PROCESS_EXIT, PROCESS_KPROBE, PROCESS_LOADER, PROCESS_LSM, PROCESS_THROTTLE, PROCESS_TRACEPOINT, PROCESS_UPROBE, RATE_LIMIT_INFO` |

### `tetragon_event_cache_fetch_retries_total`

Number of retries when fetching info from the event cache.

| label | values |
| ----- | ------ |
| `entry_type` | `ancestors_info, parent_info, pod_info, process_info` |

### `tetragon_event_cache_inserts_total`

Number of inserts to the event cache.

### `tetragon_events_exported_bytes_total`

Number of bytes exported for events

### `tetragon_events_exported_total`

Total number of events exported

### `tetragon_events_last_exported_timestamp`

Timestamp of the most recent event to be exported

### `tetragon_events_missing_process_info_total`

Number of events missing process info.

### `tetragon_export_ratelimit_events_dropped_total`

Number of events dropped on export due to rate limiting

### `tetragon_flags_total`

The total number of Tetragon flags. For internal use only.

| label | values |
| ----- | ------ |
| `type ` | `auid, clone, dataArgs, dataFilename, errorArgs, errorCWD, errorCgroupID, errorCgroupKn, errorCgroupName, errorCgroupSubsys, errorCgroupSubsysCgrp, errorCgroups, errorFilename, errorPathResolutionCwd, execve, execveat, inInitTree, miss, nocwd, procFS, rootcwd, taskWalk, truncArgs, truncFilename` |

### `tetragon_generic_kprobe_merge_errors_total`

The total number of failed attempts to merge a kprobe and kretprobe event.

| label | values |
| ----- | ------ |
| `curr_fn` | `example_kprobe` |
| `curr_type` | `enter, exit` |
| `prev_fn` | `example_kprobe` |
| `prev_type` | `enter, exit` |

### `tetragon_generic_kprobe_merge_ok_total`

The total number of successful attempts to merge a kprobe and kretprobe event.

### `tetragon_generic_kprobe_merge_pushed_total`

The total number of pushed events for later merge.

### `tetragon_handler_errors_total`

The total number of event handler errors. For internal use only.

| label | values |
| ----- | ------ |
| `error_type` | `event_handler_failed, unknown_opcode` |
| `opcode` | `0, 13, 14, 15, 16, 23, 24, 25, 26, 27, 5, 7` |

### `tetragon_handling_latency`

The latency of handling messages in us.

| label | values |
| ----- | ------ |
| `op   ` | `13, 14, 15, 16, 23, 24, 25, 26, 27, 5, 7` |

### `tetragon_map_capacity`

Capacity of a BPF map. Expected to be constant.

| label | values |
| ----- | ------ |
| `map  ` | `execve_map, tg_execve_joined_info_map` |

### `tetragon_map_entries`

The total number of in-use entries per map.

| label | values |
| ----- | ------ |
| `map  ` | `execve_map, tg_execve_joined_info_map` |

### `tetragon_map_errors_delete_total`

The number of failed deletes per map.

| label | values |
| ----- | ------ |
| `map  ` | `execve_map, tg_execve_joined_info_map` |

### `tetragon_map_errors_update_total`

The number of failed updates per map.

| label | values |
| ----- | ------ |
| `map  ` | `execve_map, tg_execve_joined_info_map` |

### `tetragon_missed_link_probes_total`

The total number of Tetragon probe missed by link.

| label | values |
| ----- | ------ |
| `attach` | `sys_panic` |
| `policy` | `monitor_panic` |

### `tetragon_missed_prog_probes_total`

The total number of Tetragon probe missed by program.

| label | values |
| ----- | ------ |
| `attach` | `sys_panic` |
| `policy` | `monitor_panic` |

### `tetragon_msg_op_total`

The total number of times we encounter a given message opcode. For internal use only.

| label | values |
| ----- | ------ |
| `msg_op` | `13, 14, 15, 16, 23, 24, 25, 26, 27, 5, 7` |

### `tetragon_notify_overflowed_events_total`

The total number of events dropped because listener buffer was full

### `tetragon_observer_ringbuf_errors_total`

Number of errors when reading Tetragon ring buffer.

### `tetragon_observer_ringbuf_events_lost_total`

Number of perf events Tetragon ring buffer lost.

### `tetragon_observer_ringbuf_events_received_total`

Number of perf events Tetragon ring buffer received.

### `tetragon_observer_ringbuf_queue_events_lost_total`

Number of perf events Tetragon ring buffer events queue lost.

### `tetragon_observer_ringbuf_queue_events_received_total`

Number of perf events Tetragon ring buffer events queue received.

### `tetragon_overhead_program_runs_total`

The total number of times BPF program was executed.

| label | values |
| ----- | ------ |
| `attach` | `sys_open` |
| `policy` | `enforce` |
| `policy_namespace` | `   ns` |
| `section` | `kprobe/sys_open` |
| `sensor` | `generic_kprobe` |

### `tetragon_overhead_program_seconds_total`

The total time of BPF program running.

| label | values |
| ----- | ------ |
| `attach` | `sys_open` |
| `policy` | `enforce` |
| `policy_namespace` | `   ns` |
| `section` | `kprobe/sys_open` |
| `sensor` | `generic_kprobe` |

### `tetragon_policyfilter_hook_container_image_missing_total`

The total number of operations when the container image was missing in the OCI hook

### `tetragon_policyfilter_hook_container_name_missing_total`

The total number of operations when the container name was missing in the OCI hook

### `tetragon_policyfilter_operations_total`

Number of policy filter operations.

| label | values |
| ----- | ------ |
| `error` | `generic-error, pod-namespace-conflict` |
| `operation` | `add, add-container, delete, update` |
| `subsys` | `pod-handlers, rthooks` |

### `tetragon_process_cache_capacity`

The capacity of the process cache. Expected to be constant.

### `tetragon_process_cache_evictions_total`

Number of process cache LRU evictions.

### `tetragon_process_cache_misses_total`

Number of process cache misses.

| label | values |
| ----- | ------ |
| `operation` | `get, remove` |

### `tetragon_process_cache_size`

The size of the process cache

### `tetragon_process_loader_stats`

Process Loader event statistics. For internal use only.

| label | values |
| ----- | ------ |
| `count` | `LoaderReceived, LoaderResolvedImm, LoaderResolvedRetry` |

### `tetragon_tracingpolicy_kernel_memory_bytes`

The amount of kernel memory in bytes used by policy's sensors non-shared BPF maps (memlock).

| label | values |
| ----- | ------ |
| `policy` | `example-tracingpolicy` |
| `policy_namespace` | `example-namespace` |

### `tetragon_tracingpolicy_loaded`

The number of loaded tracing policy by state.

| label | values |
| ----- | ------ |
| `state` | `disabled, enabled, error, load_error` |

### `tetragon_watcher_delete_pod_cache_hits`

The total hits for pod information in the deleted pod cache.

### `tetragon_watcher_errors_total`

The total number of errors for a given watcher type.

| label | values |
| ----- | ------ |
| `error` | `failed_to_get_pod` |
| `watcher` | `  k8s` |

### `tetragon_watcher_events_total`

The total number of events for a given watcher type.

| label | values |
| ----- | ------ |
| `watcher` | `  k8s` |

## Tetragon Resources Metrics

### `go_gc_duration_seconds`

A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.

### `go_gc_gogc_percent`

Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent.

### `go_gc_gomemlimit_bytes`

Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes.

### `go_goroutines`

Number of goroutines that currently exist.

### `go_info`

Information about the Go environment.

| label | values |
| ----- | ------ |
| `version` | `go1.24.5` |

### `go_memstats_alloc_bytes`

Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.

### `go_memstats_alloc_bytes_total`

Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.

### `go_memstats_buck_hash_sys_bytes`

Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.

### `go_memstats_frees_total`

Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.

### `go_memstats_gc_sys_bytes`

Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.

### `go_memstats_heap_alloc_bytes`

Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.

### `go_memstats_heap_idle_bytes`

Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.

### `go_memstats_heap_inuse_bytes`

Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes

### `go_memstats_heap_objects`

Number of currently allocated objects. Equals to /gc/heap/objects:objects.

### `go_memstats_heap_released_bytes`

Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.

### `go_memstats_heap_sys_bytes`

Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.

### `go_memstats_last_gc_time_seconds`

Number of seconds since 1970 of last garbage collection.

### `go_memstats_mallocs_total`

Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.

### `go_memstats_mcache_inuse_bytes`

Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.

### `go_memstats_mcache_sys_bytes`

Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.

### `go_memstats_mspan_inuse_bytes`

Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.

### `go_memstats_mspan_sys_bytes`

Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.

### `go_memstats_next_gc_bytes`

Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.

### `go_memstats_other_sys_bytes`

Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.

### `go_memstats_stack_inuse_bytes`

Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.

### `go_memstats_stack_sys_bytes`

Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.

### `go_memstats_sys_bytes`

Number of bytes obtained from system. Equals to /memory/classes/total:byte.

### `go_sched_gomaxprocs_threads`

The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads.

### `go_sched_latencies_seconds`

Distribution of the time goroutines have spent in the scheduler in a runnable state before actually running. Bucket counts increase monotonically. Sourced from /sched/latencies:seconds.

### `go_threads`

Number of OS threads created.

### `process_cpu_seconds_total`

Total user and system CPU time spent in seconds.

### `process_max_fds`

Maximum number of open file descriptors.

### `process_network_receive_bytes_total`

Number of bytes received by the process over the network.

### `process_network_transmit_bytes_total`

Number of bytes sent by the process over the network.

### `process_open_fds`

Number of open file descriptors.

### `process_resident_memory_bytes`

Resident memory size in bytes.

### `process_start_time_seconds`

Start time of the process since unix epoch in seconds.

### `process_virtual_memory_bytes`

Virtual memory size in bytes.

### `process_virtual_memory_max_bytes`

Maximum amount of virtual memory available in bytes.

## Tetragon Events Metrics

### `tetragon_events_total`

The total number of Tetragon events

| label | values |
| ----- | ------ |
| `binary` | `example-binary` |
| `namespace` | `example-namespace` |
| `pod  ` | `example-pod` |
| `type ` | `PROCESS_EXEC, PROCESS_EXIT, PROCESS_KPROBE, PROCESS_LOADER, PROCESS_LSM, PROCESS_THROTTLE, PROCESS_TRACEPOINT, PROCESS_UPROBE, RATE_LIMIT_INFO` |
| `workload` | `example-workload` |

### `tetragon_policy_events_total`

Policy events calls observed.

| label | values |
| ----- | ------ |
| `binary` | `example-binary` |
| `hook ` | `example_kprobe` |
| `namespace` | `example-namespace` |
| `pod  ` | `example-pod` |
| `policy` | `example-tracingpolicy` |
| `workload` | `example-workload` |

### `tetragon_syscalls_total`

System calls observed.

| label | values |
| ----- | ------ |
| `binary` | `example-binary` |
| `namespace` | `example-namespace` |
| `pod  ` | `example-pod` |
| `syscall` | `example_syscall` |
| `workload` | `example-workload` |

