package main

import gonvml "github.com/NVIDIA/go-nvml/pkg/nvml"

var brandNames = map[gonvml.BrandType]string{
	gonvml.BRAND_UNKNOWN:             "Unknown",
	gonvml.BRAND_QUADRO:              "Quadro",
	gonvml.BRAND_TESLA:               "Tesla",
	gonvml.BRAND_NVS:                 "NVS",
	gonvml.BRAND_GRID:                "GRID",
	gonvml.BRAND_GEFORCE:             "GeForce",
	gonvml.BRAND_TITAN:               "TITAN",
	gonvml.BRAND_NVIDIA_VAPPS:        "NVIDIA Virtual Applications",
	gonvml.BRAND_NVIDIA_VPC:          "NVIDIA Virtual PC",
	gonvml.BRAND_NVIDIA_VCS:          "NVIDIA Virtual Compute Server",
	gonvml.BRAND_NVIDIA_VWS:          "NVIDIA Virtual Workstation",
	gonvml.BRAND_NVIDIA_CLOUD_GAMING: "NVIDIA Cloud Gaming",
	gonvml.BRAND_QUADRO_RTX:          "Quadro RTX",
	gonvml.BRAND_NVIDIA_RTX:          "NVIDIA RTX",
	gonvml.BRAND_NVIDIA:              "NVIDIA",
	gonvml.BRAND_GEFORCE_RTX:         "GeForce RTX",
	gonvml.BRAND_TITAN_RTX:           "TITAN RTX",
}

var architectureNames = map[gonvml.DeviceArchitecture]string{
	gonvml.DEVICE_ARCH_KEPLER:    "Kepler",
	gonvml.DEVICE_ARCH_MAXWELL:   "Maxwell",
	gonvml.DEVICE_ARCH_PASCAL:    "Pascal",
	gonvml.DEVICE_ARCH_VOLTA:     "Volta",
	gonvml.DEVICE_ARCH_TURING:    "Turing",
	gonvml.DEVICE_ARCH_AMPERE:    "Ampere",
	gonvml.DEVICE_ARCH_ADA:       "Ada",
	gonvml.DEVICE_ARCH_HOPPER:    "Hopper",
	gonvml.DEVICE_ARCH_BLACKWELL: "Blackwell",
	gonvml.DEVICE_ARCH_UNKNOWN:   "Unknown",
}
