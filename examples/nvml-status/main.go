package main

import (
	"flag"
	"log"
	"os"

	gonvml "github.com/NVIDIA/go-nvml/pkg/nvml"
)

func main() {
	formatFlag := flag.String("format", "plain", "output format: plain or markdown")
	gpuID := flag.Int("id", -1, "query a specific GPU by index (0-based)")
	gpuUUID := flag.String("uuid", "", "query a specific GPU by UUID")
	flag.Parse()

	format, err := parseOutputFormat(*formatFlag)
	if err != nil {
		log.Fatal(err)
	}

	ret := gonvml.Init()
	if ret != gonvml.SUCCESS {
		log.Fatalf("nvml.Init failed: %s (%d)", gonvml.ErrorString(ret), ret)
	}
	defer func() {
		ret := gonvml.Shutdown()
		if ret != gonvml.SUCCESS {
			log.Fatalf("nvml.Shutdown failed: %s (%d)", gonvml.ErrorString(ret), ret)
		}
	}()

	count, ret := gonvml.DeviceGetCount()
	if ret != gonvml.SUCCESS {
		log.Fatalf("DeviceGetCount failed: %s (%d)", gonvml.ErrorString(ret), ret)
	}

	if *gpuID >= 0 && *gpuUUID != "" {
		log.Fatal("cannot specify both -id and -uuid")
	}
	if *gpuID >= 0 && *gpuID >= count {
		log.Fatalf("GPU id %d out of range (0-%d)", *gpuID, count-1)
	}

	w := newStatusWriter(format)
	status := collectNVMLStatus(os.Getenv("CUDA_VISIBLE_DEVICES"), count, *gpuID, *gpuUUID)
	printNVMLStatus(w, status)
}
