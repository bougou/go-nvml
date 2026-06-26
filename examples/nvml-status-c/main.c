#include <stdint.h>
#include <stdio.h>
#include <sys/types.h>
#include <unistd.h>

#include <nvml.h>

int main() {
    nvmlReturn_t ret;
    nvmlDevice_t device;
    uint32_t processSamplesCount;

    ret = nvmlInit();
    printf("nvmlInit: %s\n", nvmlErrorString(ret));

    ret = nvmlDeviceGetHandleByIndex(0, &device);
    printf("nvmlDeviceGetHandleByIndex: %s\n", nvmlErrorString(ret));

    ret =
        nvmlDeviceGetProcessUtilization(device, NULL, &processSamplesCount, 0);
    printf("nvmlDeviceGetProcessUtilization: %d, %s\n", processSamplesCount,
           nvmlErrorString(ret));

    ret = nvmlShutdown();
    printf("nvmlShutdown: %s\n", nvmlErrorString(ret));
}
