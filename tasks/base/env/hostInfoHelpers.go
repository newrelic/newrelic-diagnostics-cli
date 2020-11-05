package env

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

//Defines the struct for holding host name and OS information
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
	Processes            []Process
}

type CPU struct {
	Cores int32
	Mhz   float64
}

type Process struct {
	name string
	id   int32
}

// Name implements the genericProcess interface as part of the newrelic-cli discovery.
func (p Process) Name() (string, error) {
	return p.name, nil
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

	processInfoErr := populateHostProcessInfo(&hostInfo, ctx)
	if memInfoErr != nil {
		errorMessages = append(errorMessages, processInfoErr.Error())
	}

	if len(errorMessages) > 0 {
		return hostInfo, errors.New(strings.Join(errorMessages, "\n"))
	}

	return hostInfo, nil
}

func populateHostProcessInfo(hostInfo *HostInfo, ctx context.Context) error {
	pids, err := process.PidsWithContext(ctx)
	if err != nil {
		return err
	}

	for _, pid := range pids {
		p, err := process.NewProcess(pid)
		if err != nil {
			return fmt.Errorf("cannot read pid %d: %s", pid, err.Error())
		}

		name, err := p.Name()
		if err != nil {
			return fmt.Errorf("cannot read name of pid %d: %s", pid, err.Error())
		}

		proc := Process{
			name: name,
			id:   p.Pid,
		}

		hostInfo.Processes = append(hostInfo.Processes, proc)
	}

	return nil
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
