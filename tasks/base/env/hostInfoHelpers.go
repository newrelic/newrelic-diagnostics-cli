package env

import (
	"context"
	"errors"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"runtime"
	"strings"
)

// Defines the struct for holding host name and OS information
type HostInfo struct {
	Hostname             string
	OS                   string
	Platform             string
	PlatformFamily       string
	PlatformVersion      string
	KernelVersion        string
	KernelArch           string
	CPUs                 []CPU
	TotalVirtualMemoryMB int
}

type CPU struct {
	Cores int32
	Mhz   float64
}

type HostInfoProviderFunc func() (HostInfo, error)
type HostInfoProviderWithContextFunc func(context.Context) (HostInfo, error)

func NewHostInfo() (HostInfo, error) {
	ctx := context.Background()
	return NewHostInfoWithContext(ctx)
}

func NewHostInfoWithContext(ctx context.Context) (HostInfo, error) {
	hostInfo := HostInfo{}
	errorMessages := []string{}

	hostInfoErr := populateHostInfo(&hostInfo, ctx)

	if hostInfoErr != nil {
		errorMessages = append(errorMessages, hostInfoErr.Error())
	}

	cpuInfoErr := populateHostCPUInfo(&hostInfo, ctx)

	if cpuInfoErr != nil {
		errorMessages = append(errorMessages, cpuInfoErr.Error())
	}

	memInfoErr := populateHostMemoryInfo(&hostInfo, ctx)

	if memInfoErr != nil {
		errorMessages = append(errorMessages, memInfoErr.Error())
	}

	if len(errorMessages) > 0 {
		return hostInfo, errors.New(strings.Join(errorMessages, "\n"))
	}

	return hostInfo, nil
}

func populateHostInfo(hostInfo *HostInfo, ctx context.Context) error {

	info, err := host.InfoWithContext(ctx)
	if err != nil {
		hostInfo.OS = runtime.GOOS
		return err
	}

	hostInfo.OS = info.OS
	hostInfo.Hostname = info.Hostname
	hostInfo.Platform = info.Platform
	hostInfo.PlatformFamily = info.PlatformFamily
	hostInfo.PlatformVersion = info.PlatformVersion
	hostInfo.KernelVersion = info.KernelVersion
	hostInfo.KernelArch = info.KernelArch

	return nil
}

func populateHostCPUInfo(hostInfo *HostInfo, ctx context.Context) error {
	info, err := cpu.InfoWithContext(ctx)
	if err != nil {
		return err
	}

	for _, cpuInfo := range info {
		cpu := CPU{
			Cores: cpuInfo.Cores,
			Mhz:   cpuInfo.Mhz,
		}

		hostInfo.CPUs = append(hostInfo.CPUs, cpu)
	}

	return nil
}

func populateHostMemoryInfo(hostInfo *HostInfo, ctx context.Context) error {
	info, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return err
	}

	hostInfo.TotalVirtualMemoryMB = int((info.Total) / uint64(1048576))

	return nil
}
