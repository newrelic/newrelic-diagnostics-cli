package config

import (
	"fmt"
	"os"
	"runtime"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestInfraConfigIntegrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra/Config/* test suite")
}
func TestRegisterWithCount(t *testing.T) {
	registerWithCount := 0
	var registeredTasks []tasks.Task
	dummyRegisterWith := func(p tasks.Task, b bool) {
		registerWithCount++
		registeredTasks = append(registeredTasks, p)
	}
	type args struct {
		registrationFunc func(tasks.Task, bool)
	}

	expectedRegisteredTaskCount := 7

	tests := []struct {
		name      string
		args      args
		wantCount int
	}{
		{
			name:      "It should register six tasks",
			args:      args{dummyRegisterWith},
			wantCount: expectedRegisteredTaskCount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterWith(tt.args.registrationFunc)
			if len(registeredTasks) != tt.wantCount {
				t.Errorf("Registered tasks length didn't match registeredTasks %v, want %v", len(registeredTasks), tt.wantCount)
			}
		})
	}
}

func TestRegisterWith(t *testing.T) {
	var registeredTasks []tasks.Task
	dummyRegisterWith := func(p tasks.Task, b bool) {
		registeredTasks = append(registeredTasks, p)
	}
	type args struct {
		registrationChecker func(tasks.Task, bool)
	}

	expectedRegisteredTasks := []tasks.Task{
		InfraConfigDataDirectoryCollect{dataDirectoryGetter: getDataDir, dataDirectoryPathGetter: getDataDirPath},
		InfraConfigAgent{validationChecker: checkValidation, configChecker: checkConfig, binaryChecker: checkForBinary},
		InfraConfigIntegrationsCollect{fileFinder: tasks.FindFiles},
		InfraConfigIntegrationsValidate{fileReader: os.Open},
		InfraConfigIntegrationsMatch{runtimeOS: runtime.GOOS},
		InfraConfigIntegrationsValidateJson{},
		InfraConfigValidateJMX{mCmdExecutor: tasks.MultiCmdExecutor},
	}

	tests := []struct {
		name string
		args args
		want []tasks.Task
	}{
		{
			name: "It should register six tasks with the expected dependencies",
			args: args{dummyRegisterWith},
			want: expectedRegisteredTasks,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			RegisterWith(tt.args.registrationChecker)
			got := fmt.Sprint(registeredTasks)
			want := fmt.Sprint(tt.want)

			if got != want {
				t.Errorf("string representation of functions didn't match \n%v, \n%v", got, want)
			}
		})
	}
}
