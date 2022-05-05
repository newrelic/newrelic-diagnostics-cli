package config

import (
	"errors"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestInfraConfigDataDirectoryCollect_Explain(t *testing.T) {
	tests := []struct {
		name string
		p    InfraConfigDataDirectoryCollect
		want string
	}{
		{name: "it should return the explain", p: InfraConfigDataDirectoryCollect{}, want: "Collect New Relic Infrastructure agent data directory"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigDataDirectoryCollect{}
			if got := p.Explain(); got != tt.want {
				t.Errorf("InfraConfigDataDirectoryCollect.Explain() = \n%v, \nwant %v", got, tt.want)
			}
		})
	}
}

func TestInfraConfigDataDirectoryCollect_Identifier(t *testing.T) {
	tests := []struct {
		name string
		p    InfraConfigDataDirectoryCollect
		want tasks.Identifier
	}{
		{name: "it should return the identifier object", p: InfraConfigDataDirectoryCollect{}, want: tasks.Identifier{Name: "DataDirectoryCollect", Category: "Infra", Subcategory: "Config"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigDataDirectoryCollect{}
			if got := p.Identifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InfraConfigDataDirectoryCollect.Identifier() = \n%v, \nwant %v", got, tt.want)
			}
		})
	}
}

func TestInfraConfigDataDirectoryCollect_Dependencies(t *testing.T) {
	tests := []struct {
		name string
		p    InfraConfigDataDirectoryCollect
		want []string
	}{
		{name: "It should return the dependencies", p: InfraConfigDataDirectoryCollect{}, want: []string{"Infra/Config/Agent"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigDataDirectoryCollect{}
			if got := p.Dependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InfraConfigDataDirectoryCollect.Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfraConfigDataDirectoryCollect_Execute(t *testing.T) {
	var mockDataDir = func([]string) ([]tasks.FileCopyEnvelope, error) {
		return []tasks.FileCopyEnvelope{
			tasks.FileCopyEnvelope{Path: "/var/db/newrelic-infra/data", Identifier: "Infra/Config/DataDirectoryCollect"},
		}, nil
	}
	var mockDataDirError = func([]string) ([]tasks.FileCopyEnvelope, error) {
		return []tasks.FileCopyEnvelope{}, errors.New("no data directory detected")
	}
	var mockDataPath = func(string) []string {
		return []string{}
	}
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name string
		p    InfraConfigDataDirectoryCollect
		args args
		want tasks.Result
	}{
		{name: "It should return a None status if no Infrastructure agent detected", p: InfraConfigDataDirectoryCollect{}, args: args{}, want: tasks.Result{Summary: "No New Relic Infrastructure agent detected"}}, {name: "It should return a None status if no Infrastructure agent detected", p: InfraConfigDataDirectoryCollect{}, args: args{}, want: tasks.Result{Summary: "No New Relic Infrastructure agent detected"}},
		{name: "It should return the Infrastructure data directory", p: InfraConfigDataDirectoryCollect{dataDirectoryGetter: mockDataDir, dataDirectoryPathGetter: mockDataPath}, args: args{
			upstream: map[string]tasks.Result{"Infra/Config/Agent": tasks.Result{Status: tasks.Success}},
		},
			want: tasks.Result{
				Status:  tasks.Success,
				Summary: "New Relic Infrastructure data directory found",
				FilesToCopy: []tasks.FileCopyEnvelope{
					tasks.FileCopyEnvelope{Path: "/var/db/newrelic-infra/data", Identifier: "Infra/Config/DataDirectoryCollect"},
				},
			}},
		{name: "It should return an Error result if there is an error retrieving the data directory", p: InfraConfigDataDirectoryCollect{dataDirectoryGetter: mockDataDirError, dataDirectoryPathGetter: mockDataPath}, args: args{
			upstream: map[string]tasks.Result{"Infra/Config/Agent": tasks.Result{Status: tasks.Success}},
		},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: "Unable to get Infrastructure data directory: no data directory detected",
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			if got := p.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InfraConfigDataDirectoryCollect.Execute() = \n%v, \nwant \n%v", got, tt.want)
			}
		})
	}
}

type byPath []tasks.FileCopyEnvelope

func (p byPath) Len() int {
	return len(p)
}

func (p byPath) Less(i, j int) bool {
	return p[i].Path < p[j].Path
}

func (p byPath) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func Test_getDataDir(t *testing.T) {

	type args struct {
		paths []string
	}
	tests := []struct {
		name    string
		args    args
		want    []tasks.FileCopyEnvelope
		wantErr bool
	}{
		{name: "it should return an empty slice of file copy envelopes if no files found", args: args{}, want: []tasks.FileCopyEnvelope{}, wantErr: true},
		{name: "it should return all the files in a directory",
			args: args{[]string{"fixtures/level1dir"}},
			want: []tasks.FileCopyEnvelope{
				tasks.FileCopyEnvelope{
					Path:       filepath.FromSlash("fixtures/level1dir/level2dir/test2"),
					Identifier: "Infra/Config/DataDirectoryCollect"},
				tasks.FileCopyEnvelope{
					Path:       filepath.FromSlash("fixtures/level1dir/level2dir/test1"),
					Identifier: "Infra/Config/DataDirectoryCollect"},
				tasks.FileCopyEnvelope{
					Path:       filepath.FromSlash("fixtures/level1dir/level2dir/test3"),
					Identifier: "Infra/Config/DataDirectoryCollect"},
				tasks.FileCopyEnvelope{
					Path:       filepath.FromSlash("fixtures/level1dir/test1"),
					Identifier: "Infra/Config/DataDirectoryCollect"},
			},
		},
		
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getDataDir(tt.args.paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDataDir() error = \n%v, \nwantErr \n%v", err, tt.wantErr)
				return
			}

			sort.Sort(byPath(got))
			sort.Sort(byPath(tt.want))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDataDir() = \n%v, \nwant \n%v", got, tt.want)
			}
		})
	}
}

func Test_getDataDirPath(t *testing.T) {
	type args struct {
		osType string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "It should return on linux", args: args{osType: "linux"}, want: []string{"/var/db/newrelic-infra/data"}},
		{name: "It should return on windows", args: args{osType: "windows"}, want: []string{"C:\\Windows\\system32\\config\\systemprofile\\AppData\\Local\\New Relic", "C:\\ProgramData\\New Relic\\newrelic-infra"}},
		{name: "It should return an empty slice on non windows or linux os", args: args{osType: "tvOS"}, want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getDataDirPath(tt.args.osType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getDataDirPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
