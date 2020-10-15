package requirements

import (
	"testing"

	tasks "github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/env"
)

func TestCheckOs(t *testing.T) {

	osCheckTests := []struct {
		hostInfo env.HostInfo
		want     tasks.Status
	}{
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "",
				PlatformFamily:  "debian",
				OS:              "linux",
				Platform:        "ubuntu",
			},
			want: tasks.Warning,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "17.10",
				PlatformFamily:  "debian",
				OS:              "linux",
				Platform:        "ubuntu",
			},
			want: tasks.Success,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "17.04",
				PlatformFamily:  "debian",
				OS:              "linux",
				Platform:        "ubuntu",
			},
			want: tasks.Failure,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "42.1",
				PlatformFamily:  "suse",
				OS:              "linux",
				Platform:        "opensuse",
			},
			want: tasks.Failure,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "42.3",
				PlatformFamily:  "suse",
				OS:              "linux",
				Platform:        "opensuse",
			},
			want: tasks.Success,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "7.4",
				PlatformFamily:  "rhel",
				OS:              "linux",
				Platform:        "oracle",
			},
			want: tasks.Success,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "7.4.1708",
				PlatformFamily:  "rhel",
				OS:              "linux",
				Platform:        "centos",
			},
			want: tasks.Success,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "10.12.6",
				PlatformFamily:  "",
				OS:              "darwin",
				Platform:        "darwin",
			},
			want: tasks.Failure,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "6.3.9600 Build 9600",
				PlatformFamily:  "Server",
				OS:              "windows",
				Platform:        "Microsoft Windows Server 2012 R2 Standard",
			},
			want: tasks.Success,
		},
		{
			hostInfo: env.HostInfo{
				PlatformVersion: "5.11.44 Build 3423",
				PlatformFamily:  "Server",
				OS:              "windows",
				Platform:        "Microsoft Windows Server 2003",
			},
			want: tasks.Failure,
		},
	}

	for _, osCheckTest := range osCheckTests {

		osCheck := checkOS(osCheckTest.hostInfo)
		if osCheck.Status != osCheckTest.want {

			t.Errorf("Test failed with version %s and fam %s. Had %d wanted %d", osCheckTest.hostInfo.PlatformVersion, osCheckTest.hostInfo.PlatformFamily, osCheck.Status, osCheckTest.want)
		}
	}

}
