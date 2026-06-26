# Fix two-phase buffer allocation for process utilization info APIs

## Summary

Fix `GetProcessesUtilizationInfo` and `GetVgpuProcessesUtilizationInfo` in `pkg/nvml/device.go` to follow the NVML two-phase calling convention required by `nvmlDeviceGetProcessesUtilizationInfo` and `nvmlDeviceGetVgpuProcessesUtilizationInfo`.

The auto-generated bindings previously passed an uninitialized struct to the driver, which caused the APIs to fail at runtime.

## Problem

Both NVML APIs require the caller to:

1. Set the structure `version` field before calling.
2. Allocate the output array (`procUtilArray` / `vgpuProcUtilArray`) before fetching data.

The prior implementation did neither:

```go
func (device nvmlDevice) GetProcessesUtilizationInfo() (ProcessesUtilizationInfo, Return) {
	var processesUtilInfo ProcessesUtilizationInfo
	ret := nvmlDeviceGetProcessesUtilizationInfo(device, &processesUtilInfo)
	return processesUtilInfo, ret
}
```

This resulted in predictable failures:

| Missing requirement | NVML return code |
|---|---|
| `Version == 0` | `ERROR_ARGUMENT_VERSION_MISMATCH` |
| `ProcUtilArray == nil` | `ERROR_INSUFFICIENT_SIZE` |

`GetVgpuProcessesUtilizationInfo` had the same issue: it set `Version` but never allocated `VgpuProcUtilArray`, so it still returned `ERROR_INSUFFICIENT_SIZE` on every call.

These APIs were effectively unusable from go-nvml without manual CGO workarounds.

## Solution

Implement the same two-phase pattern already used by `GetProcessUtilization` in this package:

1. **Phase 1** — initialize the struct version and call with a `nil` output array pointer. NVML returns `ERROR_INSUFFICIENT_SIZE` and writes the required entry count into the struct.
2. **Phase 2** — allocate a slice of the required size, set the array pointer and count, then call again to populate the data.

### `GetProcessesUtilizationInfo`

```go
func (device nvmlDevice) GetProcessesUtilizationInfo() (ProcessesUtilizationInfo, Return) {
	var processesUtilInfo ProcessesUtilizationInfo
	processesUtilInfo.Version = STRUCT_VERSION(processesUtilInfo, 1)

	ret := nvmlDeviceGetProcessesUtilizationInfo(device, &processesUtilInfo)
	if ret != ERROR_INSUFFICIENT_SIZE {
		return processesUtilInfo, ret
	}

	count := processesUtilInfo.ProcessSamplesCount
	if count == 0 {
		return processesUtilInfo, ret
	}

	infos := make([]ProcessUtilizationInfo_v1, count)
	processesUtilInfo.ProcUtilArray = &infos[0]
	processesUtilInfo.ProcessSamplesCount = count

	var pinner runtime.Pinner
	pinner.Pin(&infos[0])
	defer pinner.Unpin()

	ret = nvmlDeviceGetProcessesUtilizationInfo(device, &processesUtilInfo)
	return processesUtilInfo, ret
}
```

### `GetVgpuProcessesUtilizationInfo`

Same pattern, using `VgpuProcessCount` and `VgpuProcUtilArray`, with the same `runtime.Pinner` pinning before the second CGO call.

## CGO pointer pinning

Go 1.20+ enforces that any Go pointer embedded in a struct passed to C must point to **pinned** memory. Because `ProcessesUtilizationInfo.ProcUtilArray` holds a pointer into a Go slice, the second-phase call must pin that slice with `runtime.Pinner` before invoking the NVML C API. Without pinning, the runtime panics with:

```
panic: runtime error: cgo argument has Go pointer to unpinned Go pointer
```

## API reference

Per `nvml.h`:

- `nvmlDeviceGetProcessesUtilizationInfo` — caller allocates `processSamplesCount * sizeof(nvmlProcessUtilizationInfo_t)` elements.
- `nvmlDeviceGetVgpuProcessesUtilizationInfo` — caller allocates `vgpuProcessCount * sizeof(nvmlVgpuProcessUtilizationInfo_t)` elements.

Both APIs document `ERROR_INSUFFICIENT_SIZE` when the array pointer is `NULL` or the buffer is too small, and return the required count for a follow-up call.

## Changes

| File | Change |
|---|---|
| `pkg/nvml/device.go` | Two-phase implementation for `GetProcessesUtilizationInfo` and `GetVgpuProcessesUtilizationInfo` |
| `examples/nvml-status/main.go` | Enable `GetProcessesUtilizationInfo` call in the example |

## Test plan

- [ ] Build: `go build ./pkg/nvml/...`
- [ ] On a machine with an NVIDIA GPU and driver loaded:
  - [ ] `go run ./examples/nvml-status/` — verify `GetProcessesUtilizationInfo` returns `SUCCESS` (or `ERROR_NOT_SUPPORTED` / `ERROR_NOT_FOUND` on MIG-enabled GPUs, per NVML docs)
  - [ ] Confirm `ProcessSamplesCount` matches the number of running GPU processes with non-zero utilization
  - [ ] On a vGPU-capable host, verify `GetVgpuProcessesUtilizationInfo` returns `SUCCESS` or an expected unsupported error

## Notes

- `ProcUtilArray` / `VgpuProcUtilArray` in the returned struct point to heap-allocated slices owned by the function scope; callers should read the data before discarding the returned struct, consistent with other NVML array-returning helpers in this package.
- On MIG-enabled GPUs, `nvmlDeviceGetProcessesUtilizationInfo` is documented as not currently supported.
