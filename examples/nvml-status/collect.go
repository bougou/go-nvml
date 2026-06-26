package main

import (
	"fmt"
	"unsafe"

	gonvml "github.com/NVIDIA/go-nvml/pkg/nvml"
)

func isBenignNVMLReturn(ret gonvml.Return) bool {
	return ret == gonvml.ERROR_NOT_SUPPORTED || ret == gonvml.ERROR_NOT_FOUND
}

func skippableFromRet[T any](call string, ret gonvml.Return, value T) Skippable[T] {
	switch ret {
	case gonvml.SUCCESS:
		return Skippable[T]{Call: call, OK: true, Value: value}
	case gonvml.ERROR_NOT_SUPPORTED:
		return Skippable[T]{Call: call, Reason: gonvml.ErrorString(ret)}
	case gonvml.ERROR_NOT_FOUND:
		return Skippable[T]{Call: call, Reason: gonvml.ErrorString(ret)}
	default:
		return Skippable[T]{Call: call, Reason: gonvml.ErrorString(ret)}
	}
}

func skippableSliceFromRet[T any](call string, ret gonvml.Return, items []T) SkippableSlice[T] {
	switch ret {
	case gonvml.SUCCESS:
		return SkippableSlice[T]{Call: call, OK: true, Items: items}
	case gonvml.ERROR_NOT_SUPPORTED:
		return SkippableSlice[T]{Call: call, Reason: gonvml.ErrorString(ret)}
	case gonvml.ERROR_NOT_FOUND:
		return SkippableSlice[T]{Call: call, Reason: gonvml.ErrorString(ret)}
	default:
		return SkippableSlice[T]{Call: call, Reason: gonvml.ErrorString(ret)}
	}
}

func collectNVMLStatus(cudaVisibleDevices string, count int, gpuID int, gpuUUID string) NVMLStatus {
	status := NVMLStatus{
		System:  collectSystemStatus(cudaVisibleDevices),
		Summary: SummaryStatus{GPUCount: count},
	}

	if count == 0 {
		return status
	}

	if gpuUUID != "" {
		device, ret := gonvml.DeviceGetHandleByUUID(gpuUUID)
		if ret != gonvml.SUCCESS {
			status.QueryError = skippableFromRet(fmt.Sprintf("DeviceGetHandleByUUID(%q)", gpuUUID), ret, false)
			return status
		}
		index, ret := device.GetIndex()
		if ret != gonvml.SUCCESS {
			status.QueryError = skippableFromRet("GetIndex", ret, false)
			return status
		}
		status.GPUs = []GPUStatus{collectGPUStatus(index, device)}
		return status
	}

	if gpuID >= 0 {
		device, ret := gonvml.DeviceGetHandleByIndex(gpuID)
		if ret != gonvml.SUCCESS {
			status.QueryError = skippableFromRet(fmt.Sprintf("DeviceGetHandleByIndex(%d)", gpuID), ret, false)
			return status
		}
		status.GPUs = []GPUStatus{collectGPUStatus(gpuID, device)}
		return status
	}

	status.GPUs = make([]GPUStatus, count)
	for i := 0; i < count; i++ {
		device, ret := gonvml.DeviceGetHandleByIndex(i)
		if ret != gonvml.SUCCESS {
			status.GPUs[i] = GPUStatus{
				Index:        i,
				CollectError: skippableFromRet(fmt.Sprintf("DeviceGetHandleByIndex(%d)", i), ret, false),
			}
			continue
		}
		status.GPUs[i] = collectGPUStatus(i, device)
	}
	return status
}

func collectSystemStatus(cudaVisibleDevices string) SystemStatus {
	driverVersion, ret := gonvml.SystemGetDriverVersion()
	nvmlVersion, nvmlRet := gonvml.SystemGetNVMLVersion()

	sys := SystemStatus{
		DriverVersion: skippableFromRet("SystemGetDriverVersion", ret, driverVersion),
		NVMLVersion:   skippableFromRet("SystemGetNVMLVersion", nvmlRet, nvmlVersion),
	}

	cudaVer, ret := gonvml.SystemGetCudaDriverVersion_v2()
	if ret != gonvml.SUCCESS {
		cudaVer, ret = gonvml.SystemGetCudaDriverVersion()
		sys.CudaDriverVersion = skippableFromRet("SystemGetCudaDriverVersion", ret, cudaVer)
	} else {
		sys.CudaDriverVersion = skippableFromRet("SystemGetCudaDriverVersion_v2", ret, cudaVer)
	}

	if cudaVisibleDevices == "" {
		sys.CUDAVisibleDevices = "(not set)"
	} else {
		sys.CUDAVisibleDevices = cudaVisibleDevices
	}
	return sys
}

func collectGPUStatus(index int, device gonvml.Device) GPUStatus {
	return GPUStatus{
		Index:                  index,
		Identity:               collectGPUIdentity(device),
		Capacity:               collectGPUCapacity(device),
		Usage:                  collectGPUUsage(device),
		RunningProcesses:       collectRunningProcesses(device),
		UtilizationSamples:     collectUtilizationSamples(device),
		UtilizationInfoSamples: collectUtilizationInfoSamples(device),
		MIG:                    collectMIGSection(device),
	}
}

func collectGPUIdentity(device gonvml.Device) GPUIdentity {
	name, nameRet := device.GetName()
	uuid, uuidRet := device.GetUUID()
	pciInfo, pciRet := device.GetPciInfo()
	minor, minorRet := device.GetMinorNumber()
	brand, brandRet := device.GetBrand()
	isMig, migRet := device.IsMigDeviceHandle()

	return GPUIdentity{
		Name:        skippableFromRet("GetName", nameRet, name),
		UUID:        skippableFromRet("GetUUID", uuidRet, uuid),
		PCIInfo:     skippableFromRet("GetPciInfo", pciRet, pciInfo),
		MinorNumber: skippableFromRet("GetMinorNumber", minorRet, minor),
		Brand:       skippableFromRet("GetBrand", brandRet, brand),
		IsMigDevice: skippableFromRet("IsMigDeviceHandle", migRet, isMig),
	}
}

func collectGPUCapacity(device gonvml.Device) GPUCapacity {
	arch, archRet := device.GetArchitecture()
	cores, coresRet := device.GetNumGpuCores()
	mem, memRet := device.GetMemoryInfo()
	memV2, memV2Ret := device.GetMemoryInfo_v2()

	return GPUCapacity{
		Architecture: skippableFromRet("GetArchitecture", archRet, arch),
		GpuCores:     skippableFromRet("GetNumGpuCores", coresRet, cores),
		MemoryV1:     skippableFromRet("GetMemoryInfo", memRet, mem),
		MemoryV2:     skippableFromRet("GetMemoryInfo_v2", memV2Ret, memV2),
	}
}

func collectDeviceAttributes(device gonvml.Device) Skippable[gonvml.DeviceAttributes] {
	attrs, ret := device.GetAttributes()
	return skippableFromRet("GetAttributes", ret, attrs)
}

func collectMemoryV1(device gonvml.Device) Skippable[gonvml.Memory] {
	mem, ret := device.GetMemoryInfo()
	return skippableFromRet("GetMemoryInfo", ret, mem)
}

func collectMemoryV2(device gonvml.Device) Skippable[gonvml.Memory_v2] {
	mem, ret := device.GetMemoryInfo_v2()
	return skippableFromRet("GetMemoryInfo_v2", ret, mem)
}

func collectGPUUsage(device gonvml.Device) GPUUsage {
	util, utilRet := device.GetUtilizationRates()
	powerUsage, powerRet := device.GetPowerUsage()
	powerLimit, limitRet := device.GetPowerManagementLimit()
	temp, tempRet := device.GetTemperature(gonvml.TEMPERATURE_GPU)
	fanSpeed, fanRet := device.GetFanSpeed()
	pState, pStateRet := device.GetPerformanceState()

	return GPUUsage{
		UtilizationRates: skippableFromRet("GetUtilizationRates", utilRet, util),
		MemoryV1:         collectMemoryV1(device),
		MemoryV2:         collectMemoryV2(device),
		PowerUsage:       skippableFromRet("GetPowerUsage", powerRet, powerUsage),
		PowerLimit:       skippableFromRet("GetPowerManagementLimit", limitRet, powerLimit),
		Temperature:      skippableFromRet("GetTemperature", tempRet, temp),
		FanSpeed:         skippableFromRet("GetFanSpeed", fanRet, fanSpeed),
		PerformanceState: skippableFromRet("GetPerformanceState", pStateRet, pState),
	}
}

func collectRunningProcesses(device gonvml.Device) SkippableSlice[gonvml.ProcessInfo] {
	processes, ret := device.GetComputeRunningProcesses()
	return skippableSliceFromRet("GetComputeRunningProcesses", ret, processes)
}

func collectUtilizationSamples(device gonvml.Device) SkippableSlice[gonvml.ProcessUtilizationSample] {
	samples, ret := device.GetProcessUtilization(0)
	return skippableSliceFromRet("GetProcessUtilization", ret, samples)
}

func collectUtilizationInfoSamples(device gonvml.Device) SkippableSlice[gonvml.ProcessUtilizationInfo_v1] {
	info, ret := device.GetProcessesUtilizationInfo()
	var items []gonvml.ProcessUtilizationInfo_v1
	if ret == gonvml.SUCCESS {
		count := int(info.ProcessSamplesCount)
		if info.ProcUtilArray != nil && count > 0 {
			items = unsafe.Slice(info.ProcUtilArray, count)
		}
	}
	return skippableSliceFromRet("GetProcessesUtilizationInfo", ret, items)
}

func collectMIGSection(device gonvml.Device) *MIGSection {
	isMigDevice, ret := device.IsMigDeviceHandle()
	if ret != gonvml.SUCCESS {
		if isBenignNVMLReturn(ret) {
			return nil
		}
		return &MIGSection{Error: skippableFromRet("IsMigDeviceHandle", ret, false)}
	}
	if isMigDevice {
		return nil
	}

	currentMode, pendingMode, ret := device.GetMigMode()
	if ret != gonvml.SUCCESS {
		if isBenignNVMLReturn(ret) {
			return nil
		}
		return &MIGSection{Error: skippableFromRet("GetMigMode", ret, false)}
	}

	if currentMode == 0 {
		return &MIGSection{Disabled: true}
	}

	migCount, ret := device.GetMaxMigDeviceCount()
	if ret != gonvml.SUCCESS {
		section := &MIGSection{
			CurrentMode: currentMode,
			PendingMode: pendingMode,
		}
		if !isBenignNVMLReturn(ret) {
			section.Error = skippableFromRet("GetMaxMigDeviceCount", ret, false)
		}
		return section
	}

	section := &MIGSection{
		CurrentMode: currentMode,
		PendingMode: pendingMode,
		MaxSlots:    migCount,
	}

	for j := 0; j < migCount; j++ {
		migDevice, ret := device.GetMigDeviceHandleByIndex(j)
		if ret == gonvml.SUCCESS {
			section.ActiveCount++
			section.Devices = append(section.Devices, collectMIGDeviceStatus(j, migDevice))
		}
	}

	return section
}

func collectMIGDeviceStatus(index int, migDevice gonvml.Device) MIGDeviceStatus {
	uuid, uuidRet := migDevice.GetUUID()
	gi, giRet := migDevice.GetGpuInstanceId()
	ci, ciRet := migDevice.GetComputeInstanceId()

	return MIGDeviceStatus{
		Index:             index,
		UUID:              skippableFromRet("GetUUID", uuidRet, uuid),
		GPUInstanceID:     skippableFromRet("GetGpuInstanceId", giRet, gi),
		ComputeInstanceID: skippableFromRet("GetComputeInstanceId", ciRet, ci),
		Attributes:        collectDeviceAttributes(migDevice),
		MemoryV1:          collectMemoryV1(migDevice),
		MemoryV2:          collectMemoryV2(migDevice),
		RunningProcesses:  collectRunningProcesses(migDevice),
	}
}
