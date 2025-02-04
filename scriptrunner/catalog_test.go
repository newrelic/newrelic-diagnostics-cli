package scriptrunner

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
)

func TestCatalog_GetCatalog(t *testing.T) {
	type fields struct {
		Deps ICatalogDependencies
	}
	tests := []struct {
		name    string
		fields  fields
		want    []CatalogItem
		wantErr bool
	}{
		{
			name:   "Successful response from gh",
			fields: fields{Deps: &mocks.MockCatalogDependenciesSuccessful{}},
			want: []CatalogItem{
				{
					Name:        "test",
					Filename:    "test.sh",
					Description: "test\ntest\ntest",
					Type:        "bash",
					OS:          "darwin",
					OutputFiles: []string{"file*.log", "*file.log"},
				},
			},
			wantErr: false,
		},
		{
			name:    "Error response from gh for list of files",
			fields:  fields{Deps: &mocks.MockCatalogDependenciesErrorList{}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error response from gh for yml",
			fields:  fields{Deps: &mocks.MockCatalogDependenciesErrorFile{}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Catalog{
				Deps: tt.fields.Deps,
			}

			got, err := c.GetCatalog()
			if (err != nil) != tt.wantErr {
				t.Errorf("ScriptCatalog.GetCatalog() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ScriptCatalog.GetCatalog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCatalog_downloadFile(t *testing.T) {
	type fields struct {
		Deps ICatalogDependencies
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name:    "Rate Limit test",
			fields:  fields{Deps: &mocks.MockCatalogDependencies403Error{}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Catalog{
				Deps: tt.fields.Deps,
			}
			got, err := c.downloadFile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Catalog.downloadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Catalog.downloadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
