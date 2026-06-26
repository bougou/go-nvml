package main

import gonvml "github.com/NVIDIA/go-nvml/pkg/nvml"

const invalidInstanceID = 0xFFFFFFFF

// Skippable holds a value collected from NVML, or a skip reason for unsupported calls.
type Skippable[T any] struct {
	Value  T
	OK     bool
	Call   string
	Reason string
}

// SkippableSlice holds a slice collected from NVML, or a skip reason.
type SkippableSlice[T any] struct {
	Items  []T
	OK     bool
	Call   string
	Reason string
}

// NVMLStatus is the complete snapshot of node GPU status.
type NVMLStatus struct {
	System     SystemStatus
	Summary    SummaryStatus
	QueryError Skippable[bool]
	GPUs       []GPUStatus
}

type SystemStatus struct {
	DriverVersion      Skippable[string]
	NVMLVersion        Skippable[string]
	CudaDriverVersion  Skippable[int]
	CUDAVisibleDevices string
}

type SummaryStatus struct {
	GPUCount int
}

type GPUStatus struct {
	Index                  int
	CollectError           Skippable[bool]
	Identity               GPUIdentity
	Capacity               GPUCapacity
	Usage                  GPUUsage
	RunningProcesses       SkippableSlice[gonvml.ProcessInfo]
	UtilizationSamples     SkippableSlice[gonvml.ProcessUtilizationSample]
	UtilizationInfoSamples SkippableSlice[gonvml.ProcessUtilizationInfo_v1]
	MIG                    *MIGSection
}

type GPUIdentity struct {
	Name        Skippable[string]
	UUID        Skippable[string]
	PCIInfo     Skippable[gonvml.PciInfo]
	MinorNumber Skippable[int]
	Brand       Skippable[gonvml.BrandType]
	IsMigDevice Skippable[bool]
}

type GPUCapacity struct {
	Architecture Skippable[gonvml.DeviceArchitecture]
	GpuCores     Skippable[int]
	MemoryV1     Skippable[gonvml.Memory]
	MemoryV2     Skippable[gonvml.Memory_v2]
}

type GPUUsage struct {
	UtilizationRates Skippable[gonvml.Utilization]
	MemoryV1         Skippable[gonvml.Memory]
	MemoryV2         Skippable[gonvml.Memory_v2]
	PowerUsage       Skippable[uint32]
	PowerLimit       Skippable[uint32]
	Temperature      Skippable[uint32]
	FanSpeed         Skippable[uint32]
	PerformanceState Skippable[gonvml.Pstates]
}

type MIGSection struct {
	Disabled    bool
	CurrentMode int
	PendingMode int
	ActiveCount int
	MaxSlots    int
	Devices     []MIGDeviceStatus
	Error       Skippable[bool]
}

type MIGDeviceStatus struct {
	Index             int
	UUID              Skippable[string]
	GPUInstanceID     Skippable[int]
	ComputeInstanceID Skippable[int]
	Attributes        Skippable[gonvml.DeviceAttributes]
	MemoryV1          Skippable[gonvml.Memory]
	MemoryV2          Skippable[gonvml.Memory_v2]
	RunningProcesses  SkippableSlice[gonvml.ProcessInfo]
}
