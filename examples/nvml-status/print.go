package main

import (
	"fmt"

	gonvml "github.com/NVIDIA/go-nvml/pkg/nvml"
)

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	value := float64(b) / float64(div)
	return fmt.Sprintf("%.1f %ciB", value, "KMGTPE"[exp])
}

func formatPercent(value uint32) string {
	return fmt.Sprintf("%d%%", value)
}

func formatMilliwatts(mw uint32) string {
	return fmt.Sprintf("%.1f W", float64(mw)/1000.0)
}

func formatBrand(brand gonvml.BrandType) string {
	name, ok := brandNames[brand]
	if !ok {
		return fmt.Sprintf("unknown (%d)", brand)
	}
	return fmt.Sprintf("%s (%d)", name, brand)
}

func int8SliceToString(b []int8) string {
	n := 0
	for n < len(b) && b[n] != 0 {
		n++
	}
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = byte(b[i])
	}
	return string(buf)
}

func formatPCIBusID(pci gonvml.PciInfo) string {
	return int8SliceToString(pci.BusId[:])
}

func formatArchitecture(arch gonvml.DeviceArchitecture) string {
	name, ok := architectureNames[arch]
	if !ok {
		return fmt.Sprintf("unknown (%d)", arch)
	}
	return fmt.Sprintf("%s (%d)", name, arch)
}

func memoryUsedPercent(used, total uint64) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", float64(used)*100/float64(total))
}

func printSkippable[T any](w statusWriter, key string, field Skippable[T], format func(T) string) {
	if field.OK {
		w.Field(key, format(field.Value))
	} else if field.Reason != "" {
		w.Field(key, formatSkipped(field.Call, field.Reason))
	}
}

func printNVMLStatus(w statusWriter, status NVMLStatus) {
	w.H1("GPU Node Status")
	w.Blank()

	printSystemStatus(w, status.System)
	printSummaryStatus(w, status.Summary)

	if status.QueryError.Reason != "" {
		w.Text(formatSkipped(status.QueryError.Call, status.QueryError.Reason))
		w.Blank()
	}

	if status.Summary.GPUCount == 0 {
		w.Text("No GPUs found.")
		return
	}

	for _, gpu := range status.GPUs {
		printGPUStatus(w, gpu)
	}
}

func printSystemStatus(w statusWriter, sys SystemStatus) {
	w.H2("System")
	printSkippable(w, "Driver version", sys.DriverVersion, func(v string) string { return v })
	printSkippable(w, "NVML version", sys.NVMLVersion, func(v string) string { return v })
	printSkippable(w, "CUDA driver version", sys.CudaDriverVersion, func(v int) string {
		return fmt.Sprintf("%d", v)
	})
	w.Field("CUDA_VISIBLE_DEVICES", sys.CUDAVisibleDevices)
	w.Blank()
}

func printSummaryStatus(w statusWriter, summary SummaryStatus) {
	w.H2("Summary")
	w.Field("GPU count", fmt.Sprintf("%d", summary.GPUCount))
	w.Blank()
}

func printGPUStatus(w statusWriter, gpu GPUStatus) {
	w.Rule()
	w.H2(fmt.Sprintf("GPU %d", gpu.Index))
	w.Blank()

	if gpu.CollectError.Reason != "" {
		w.Text(formatSkipped(gpu.CollectError.Call, gpu.CollectError.Reason))
		w.Blank()
		return
	}

	printGPUIdentity(w, gpu.Identity)
	printGPUCapacity(w, gpu.Capacity)
	printGPUUsage(w, gpu.Usage)
	printRunningProcessesStatus(w, gpu.RunningProcesses)
	printUtilizationSamplesStatus(w, gpu.UtilizationSamples)
	printUtilizationInfoSamplesStatus(w, gpu.UtilizationInfoSamples)
	if gpu.MIG != nil {
		printMIGSectionStatus(w, *gpu.MIG)
	}
	w.Blank()
}

func printGPUIdentity(w statusWriter, identity GPUIdentity) {
	w.H3("Identity")
	printSkippable(w, "Name", identity.Name, func(v string) string { return v })
	printSkippable(w, "UUID", identity.UUID, func(v string) string { return v })
	printSkippable(w, "PCI bus ID", identity.PCIInfo, formatPCIBusID)
	printSkippable(w, "Minor number", identity.MinorNumber, func(v int) string { return fmt.Sprintf("%d", v) })
	printSkippable(w, "Brand", identity.Brand, formatBrand)
	printSkippable(w, "MIG device handle", identity.IsMigDevice, func(v bool) string {
		return fmt.Sprintf("%v", v)
	})
	w.Blank()
}

func printGPUCapacity(w statusWriter, capacity GPUCapacity) {
	w.H3("Capacity")
	printSkippable(w, "Architecture", capacity.Architecture, formatArchitecture)
	printSkippable(w, "GPU cores", capacity.GpuCores, func(v int) string { return fmt.Sprintf("%d", v) })
	printSkippable(w, "VRAM total (v1)", capacity.MemoryV1, func(v gonvml.Memory) string { return formatMiB(v.Total) })
	printSkippable(w, "VRAM total (v2)", capacity.MemoryV2, func(v gonvml.Memory_v2) string { return formatMiB(v.Total) })
	w.Blank()
}

func printMIGDeviceAttributes(w statusWriter, attrs Skippable[gonvml.DeviceAttributes]) {
	w.Group("Attributes")
	defer w.EndSection()

	if attrs.OK {
		a := attrs.Value
		w.Field("Multiprocessors", fmt.Sprintf("%d", a.MultiprocessorCount))
		w.Field("Shared copy engines", fmt.Sprintf("%d", a.SharedCopyEngineCount))
		w.Field("Shared decoders", fmt.Sprintf("%d", a.SharedDecoderCount))
		w.Field("Shared encoders", fmt.Sprintf("%d", a.SharedEncoderCount))
		w.Field("Shared JPEG engines", fmt.Sprintf("%d", a.SharedJpegCount))
		w.Field("Shared OFA engines", fmt.Sprintf("%d", a.SharedOfaCount))
		w.Field("GPU instance slices", fmt.Sprintf("%d", a.GpuInstanceSliceCount))
		w.Field("Compute instance slices", fmt.Sprintf("%d", a.ComputeInstanceSliceCount))
		w.Field("Memory", formatMiB(a.MemorySizeMB * bytesPerMiB))
	} else if attrs.Reason != "" {
		w.Text(formatSkipped(attrs.Call, attrs.Reason))
	}
}

func printMemoryV1(w statusWriter, mem Skippable[gonvml.Memory]) {
	w.Group("VRAM (v1)")
	defer w.EndSection()

	if mem.OK {
		w.Field("Used", formatMemoryUsage(mem.Value.Used, mem.Value.Total))
		w.Field("Free", formatMiB(mem.Value.Free))
	} else if mem.Reason != "" {
		w.Text(formatSkipped(mem.Call, mem.Reason))
	}
}

func printMemoryV2(w statusWriter, mem Skippable[gonvml.Memory_v2]) {
	w.Group("VRAM (v2)")
	defer w.EndSection()

	if mem.OK {
		w.Field("Used", formatMemoryUsage(mem.Value.Used, mem.Value.Total))
		w.Field("Reserved", formatMiB(mem.Value.Reserved))
		w.Field("Free", formatMiB(mem.Value.Free))
	} else if mem.Reason != "" {
		w.Text(formatSkipped(mem.Call, mem.Reason))
	}
}

func printGPUUsage(w statusWriter, usage GPUUsage) {
	w.H3("Usage")
	if usage.UtilizationRates.OK {
		w.Field("GPU utilization", formatPercent(usage.UtilizationRates.Value.Gpu))
		w.Field("Memory controller utilization", formatPercent(usage.UtilizationRates.Value.Memory))
	} else if usage.UtilizationRates.Reason != "" {
		skip := formatSkipped(usage.UtilizationRates.Call, usage.UtilizationRates.Reason)
		w.Field("GPU utilization", skip)
		w.Field("Memory controller utilization", skip)
	}
	printMemoryV1(w, usage.MemoryV1)
	printMemoryV2(w, usage.MemoryV2)

	if usage.PowerUsage.OK {
		powerText := formatMilliwatts(usage.PowerUsage.Value)
		if usage.PowerLimit.OK {
			powerText += " (limit " + formatMilliwatts(usage.PowerLimit.Value) + ")"
		}
		w.Field("Power draw", powerText)
	} else if usage.PowerUsage.Reason != "" {
		w.Field("Power draw", formatSkipped(usage.PowerUsage.Call, usage.PowerUsage.Reason))
	}

	printSkippable(w, "Temperature", usage.Temperature, func(v uint32) string {
		return fmt.Sprintf("%d C", v)
	})
	printSkippable(w, "Fan speed", usage.FanSpeed, formatPercent)
	printSkippable(w, "Performance state", usage.PerformanceState, func(v gonvml.Pstates) string {
		return fmt.Sprintf("P%d", v)
	})
	w.Blank()
}

func printRunningProcessesStatus(w statusWriter, processes SkippableSlice[gonvml.ProcessInfo]) {
	if !processes.OK {
		if processes.Reason != "" {
			w.H3("Running processes")
			w.Text(formatSkipped(processes.Call, processes.Reason))
			w.Blank()
		}
		return
	}

	w.H3(fmt.Sprintf("Running processes (%d)", len(processes.Items)))
	if len(processes.Items) == 0 {
		w.Text("(none)")
	} else {
		for _, proc := range processes.Items {
			printRunningProcess(w, proc)
		}
	}
	w.Blank()
}

func printRunningProcess(w statusWriter, proc gonvml.ProcessInfo) {
	w.H4(fmt.Sprintf("PID %d", proc.Pid))
	w.Field("VRAM", formatMiB(proc.UsedGpuMemory))
	if proc.GpuInstanceId != invalidInstanceID {
		w.Field("GPU instance ID", fmt.Sprintf("%d", proc.GpuInstanceId))
	}
	if proc.ComputeInstanceId != invalidInstanceID {
		w.Field("Compute instance ID", fmt.Sprintf("%d", proc.ComputeInstanceId))
	}
}

func printUtilizationSamplesStatus(w statusWriter, samples SkippableSlice[gonvml.ProcessUtilizationSample]) {
	if !samples.OK {
		if samples.Reason != "" {
			w.H3("Utilization samples")
			w.Text(formatSkipped(samples.Call, samples.Reason))
			w.Blank()
		}
		return
	}

	w.H3(fmt.Sprintf("Utilization samples (%d)", len(samples.Items)))
	if len(samples.Items) == 0 {
		w.Text("(none)")
	} else {
		for _, sample := range samples.Items {
			w.H4(fmt.Sprintf("PID %d", sample.Pid))
			w.Field("SM", formatPercent(sample.SmUtil))
			w.Field("Memory", formatPercent(sample.MemUtil))
			w.Field("Encoder", formatPercent(sample.EncUtil))
			w.Field("Decoder", formatPercent(sample.DecUtil))
		}
	}
	w.Blank()
}

func printUtilizationInfoSamplesStatus(w statusWriter, samples SkippableSlice[gonvml.ProcessUtilizationInfo_v1]) {
	if !samples.OK {
		if samples.Reason != "" {
			w.H3("Utilization info samples")
			w.Text(formatSkipped(samples.Call, samples.Reason))
			w.Blank()
		}
		return
	}

	w.H3(fmt.Sprintf("Utilization info samples (%d)", len(samples.Items)))
	if len(samples.Items) == 0 {
		w.Text("(none)")
	} else {
		for _, sample := range samples.Items {
			w.H4(fmt.Sprintf("PID %d", sample.Pid))
			w.Field("SM", formatPercent(sample.SmUtil))
			w.Field("Memory", formatPercent(sample.MemUtil))
			w.Field("Encoder", formatPercent(sample.EncUtil))
			w.Field("Decoder", formatPercent(sample.DecUtil))
			w.Field("JPEG", formatPercent(sample.JpgUtil))
			w.Field("OFA", formatPercent(sample.OfaUtil))
		}
	}
	w.Blank()
}

func printMIGSectionStatus(w statusWriter, mig MIGSection) {
	w.H3("MIG")
	if mig.Error.Reason != "" {
		w.Text(formatSkipped(mig.Error.Call, mig.Error.Reason))
		w.Blank()
		return
	}
	if mig.Disabled {
		w.Text("Disabled")
		w.Blank()
		return
	}

	w.Field("Mode", fmt.Sprintf("enabled (current %d, pending %d)", mig.CurrentMode, mig.PendingMode))
	if mig.MaxSlots > 0 {
		w.Field("MIG devices", fmt.Sprintf("%d active / %d max slots", mig.ActiveCount, mig.MaxSlots))
	}

	for _, device := range mig.Devices {
		w.H4(fmt.Sprintf("MIG %d", device.Index))
		printSkippable(w, "UUID", device.UUID, func(v string) string { return v })
		printSkippable(w, "GPU instance ID", device.GPUInstanceID, func(v int) string {
			return fmt.Sprintf("%d", v)
		})
		printSkippable(w, "Compute instance ID", device.ComputeInstanceID, func(v int) string {
			return fmt.Sprintf("%d", v)
		})
		printMIGDeviceAttributes(w, device.Attributes)
		printMemoryV1(w, device.MemoryV1)
		printMemoryV2(w, device.MemoryV2)

		if device.RunningProcesses.OK && len(device.RunningProcesses.Items) > 0 {
			w.Field("Running processes", fmt.Sprintf("%d", len(device.RunningProcesses.Items)))
			for _, proc := range device.RunningProcesses.Items {
				w.H4(fmt.Sprintf("PID %d", proc.Pid))
				w.Field("VRAM", formatMiB(proc.UsedGpuMemory))
			}
		}
	}

	w.Blank()
}
