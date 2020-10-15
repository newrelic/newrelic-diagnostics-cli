package tasks

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestVersionIsCompatible(t *testing.T) {
	type args struct {
		version      string
		requirements []string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Non integer input throws an error",
			args: args{
				version:      "nine",
				requirements: []string{"9.0"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Single requirement starred case is correctly handled",
			args: args{
				version:      "9.0.1",
				requirements: []string{"9.0.*"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Single requirement plus positive case is correctly handled",
			args: args{
				version:      "11.0",
				requirements: []string{"9.0+"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Single requirement plus negative case is correctly handled",
			args: args{
				version:      "9.0",
				requirements: []string{"9.1+"},
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Single requirement positive case is correctly handled",
			args: args{
				version:      "11.0",
				requirements: []string{"11"},
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Single requirement negative case is correctly handled",
			args: args{
				version:      "11.0",
				requirements: []string{"9.0"},
			},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VersionIsCompatible(tt.args.version, tt.args.requirements)
			if (err != nil) != tt.wantErr {
				t.Errorf("VersionIsCompatible() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("VersionIsCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVer_IsGreaterThanEq(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
		Build int
	}
	type args struct {
		min Ver
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Single component version passes vs identical min version",
			fields: fields{1, 0, 0, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs identical min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 2, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs identical min version",
			fields: fields{1, 2, 3, 0},
			args: args{
				Ver{1, 2, 3, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs identical min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 3, 4},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs smaller single-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs smaller two-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs smaller three-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 1, 700, 0},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs smaller four-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 1, 256, 60000},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs smaller single-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs smaller two-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs smaller three-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 1, 700, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs smaller four-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 1, 256, 60000},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs smaller single-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs smaller two-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs smaller three-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 1, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs smaller four-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 1, 60000},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs smaller single-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs smaller two-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs smaller three-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 2, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs smaller four-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 3, 3},
			},
			want: true,
		},
		// All zeroes cases (Zero min ver used when all versions supported)
		{
			name:   "Single component version passes vs all-zeroes min version",
			fields: fields{1, 0, 0, 0},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs all-zeroes min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs all-zeroes min version",
			fields: fields{1, 2, 3, 0},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs all-zeroes min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: true,
		},
		// Failure cases
		{
			name:   "Single-component version fails vs larger single-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{3, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Single-component version fails vs larger two-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{2, 1, 0, 0},
			},
			want: false,
		},
		{
			name:   "Single-component version fails vs larger three-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{2, 0, 1, 0},
			},
			want: false,
		},
		{
			name:   "Single-component version fails vs larger four-component min version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{2, 1, 256, 60000},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs larger single-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{2, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs larger two-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 3, 0, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs larger three-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 2, 3, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs larger four-component min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 2, 0, 3},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs larger single-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{2, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs larger two-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 3, 0, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs larger three-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 3, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs larger four-component min version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 2, 3},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs larger single-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{2, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs larger two-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 3, 0, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs larger three-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 4, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs larger four-component min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 3, 5},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Ver{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
				Build: tt.fields.Build,
			}
			if got := v.IsGreaterThanEq(tt.args.min); got != tt.want {
				t.Errorf("ver.IsGreaterThanEq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVer_IsLessThanEq(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
		Build int
	}
	type args struct {
		max Ver
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Single component version passes vs identical min version",
			fields: fields{1, 0, 0, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs identical min version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 2, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs identical min version",
			fields: fields{1, 2, 3, 0},
			args: args{
				Ver{1, 2, 3, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs identical min version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 3, 4},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs larger single-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{3, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs larger two-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{3, 1, 0, 0},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs larger three-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{3, 1, 700, 0},
			},
			want: true,
		},
		{
			name:   "Single-component version passes vs larger four-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{3, 1, 256, 60000},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs larger single-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{2, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs larger two-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 3, 0, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs larger three-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 2, 3, 0},
			},
			want: true,
		},
		{
			name:   "Two component version passes vs larger four-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 2, 0, 1},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs larger single-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{2, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs larger two-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 3, 0, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs larger three-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 3, 0},
			},
			want: true,
		},
		{
			name:   "Three component version passes vs larger four-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 2, 3},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs larger single-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{2, 0, 0, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs larger two-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 3, 0, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs larger three-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 4, 0},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs larger four-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 3, 5},
			},
			want: true,
		},
		// failure cases
		{
			name:   "Single-component version fails vs smaller single-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Single-component version fails vs smaller two-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: false,
		},
		{
			name:   "Single-component version fails vs smaller three-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 1, 700, 0},
			},
			want: false,
		},
		{
			name:   "Single-component version fails vs smaller four-component max version",
			fields: fields{2, 0, 0, 0},
			args: args{
				Ver{1, 1, 256, 60000},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs smaller single-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs smaller two-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs smaller three-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 1, 700, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs smaller four-component max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{1, 1, 256, 60000},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs smaller single-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs smaller two-component max version",
			fields: fields{1, 2, 1, 0},
			args: args{
				Ver{1, 2, 0, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs smaller three-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 1, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs smaller four-component max version",
			fields: fields{1, 2, 2, 0},
			args: args{
				Ver{1, 2, 1, 60000},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs smaller single-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs smaller two-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 1, 0, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs smaller three-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 2, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs smaller four-component max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{1, 2, 3, 3},
			},
			want: false,
		},
		// All zeroes cases (Zero max ver used when all versions supported)
		{
			name:   "Single component version fails vs all-zeroes max version",
			fields: fields{1, 0, 0, 0},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Two component version fails vs all-zeroes max version",
			fields: fields{1, 2, 0, 0},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Three component version fails vs all-zeroes max version",
			fields: fields{1, 2, 3, 0},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: false,
		},
		{
			name:   "Four component version fails vs all-zeroes max version",
			fields: fields{1, 2, 3, 4},
			args: args{
				Ver{0, 0, 0, 0},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Ver{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
				Build: tt.fields.Build,
			}
			if got := v.IsLessThanEq(tt.args.max); got != tt.want {
				t.Errorf("ver.IsLessThanEq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ParseVersion(t *testing.T) {
	type args struct {
		version string
	}
	tests := []struct {
		name    string
		args    args
		want    Ver
		wantErr bool
	}{
		{
			name: "Infinity special case version is correctly parsed",
			args: args{
				version: "infinity",
			},
			want: Ver{
				Major: int(^uint(0) >> 1),
				Minor: int(^uint(0) >> 1),
				Build: int(^uint(0) >> 1),
				Patch: int(^uint(0) >> 1),
			},
		},
		{
			name: "Single component version is correctly parsed",
			args: args{
				version: "4",
			},
			want: Ver{4, 0, 0, 0},
		},
		{
			name: "Two component version is correctly parsed",
			args: args{
				version: "4.1",
			},
			want: Ver{4, 1, 0, 0},
		},
		{
			name: "Three component version is correctly parsed",
			args: args{
				version: "4.1.120",
			},
			want: Ver{4, 1, 120, 0},
		},
		{
			name: "Four component version is correctly parsed",
			args: args{
				version: "4.1.120.1",
			},
			want: Ver{4, 1, 120, 1},
		},
		{
			name: "Single component version with spaces is correctly parsed",
			args: args{
				version: " 4  ",
			},
			want: Ver{4, 0, 0, 0},
		},
		{
			name: "Two component version with spaces is correctly parsed",
			args: args{
				version: " 4.1  ",
			},
			want: Ver{4, 1, 0, 0},
		},
		{
			name: "Three component version with spaces is correctly parsed",
			args: args{
				version: "  4.1.120   ",
			},
			want: Ver{4, 1, 120, 0},
		},
		{
			name: "Four component version with spaces is correctly parsed",
			args: args{
				version: "  4.1.120.1  ",
			},
			want: Ver{4, 1, 120, 1},
		},
		{
			name: "Garbage version throws error",
			args: args{
				version: "Error! Unable to frobnosticate secondary whatzits",
			},
			want:    Ver{0, 0, 0, 0},
			wantErr: true,
		},
		{
			name: "Abitrary chars in version throws error and returns zeroed version",
			args: args{
				version: "4.0b1",
			},
			want:    Ver{0, 0, 0, 0},
			wantErr: true,
		},
		{
			name: "Single component plus * version is correctly parsed",
			args: args{
				version: "4.*",
			},
			want: Ver{
				Major: 4,
				Minor: int(^uint(0) >> 1),
				Patch: 0,
				Build: 0,
			},
			wantErr: false,
		},
		{
			name: "Two component star * version is correctly parsed",
			args: args{
				version: "4.1.*",
			},
			want: Ver{
				Major: 4,
				Minor: 1,
				Patch: int(^uint(0) >> 1),
				Build: 0,
			},
			wantErr: false,
		},
		{
			name: "Three component star * version is correctly parsed",
			args: args{
				version: "4.1.2.*",
			},
			want: Ver{
				Major: 4,
				Minor: 1,
				Patch: 2,
				Build: int(^uint(0) >> 1),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVersion(tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVer_CheckCompatibility(t *testing.T) {
	type fields struct {
		Major int
		Minor int
		Patch int
		Build int
	}
	type args struct {
		requirements []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "Major version wildcard operator matches 3 component version",
			fields: fields{
				Major: 9,
				Minor: 2,
				Patch: 7,
			},
			args: args{
				requirements: []string{"2-*"},
			},
			want: true,
		},
		{
			name:   "Minor version wildcard operator matches 4 component version",
			fields: fields{7, 2, 7, 11},
			args: args{
				requirements: []string{"6.1-7.*"},
			},
			want: true,
		},
		{
			name:   "Patch version wildcard operator matches 4 component version",
			fields: fields{7, 2, 7, 11},
			args: args{
				requirements: []string{"6.1-7.2.*"},
			},
			want: true,
		},
		{
			name:   "Build version wildcard operator matches 4 component version",
			fields: fields{7, 2, 7, 11},
			args: args{
				requirements: []string{"6.1-7.2.7.*"},
			},
			want: true,
		},
		{
			name:   "Minor version wildcard operator fails if major version too high",
			fields: fields{7, 2, 7, 11},
			args: args{
				requirements: []string{"6.1-6.*"},
			},
			want: false,
		},
		{
			name:   "Patch version wildcard operator fails if minor version too high",
			fields: fields{7, 2, 7, 11},
			args: args{
				requirements: []string{"6.1-7.1.*"},
			},
			want: false,
		},
		{
			name:   "Build version wildcard operator fails if patch version too high",
			fields: fields{7, 2, 7, 11},
			args: args{
				requirements: []string{"6.1-7.2.5.*"},
			},
			want: false,
		},
		{
			name: "Version matches explicit standalone requirement",
			fields: fields{
				Major: 9,
				Minor: 0,
			},
			args: args{
				requirements: []string{"9.0"},
			},
			want: true,
		},
		{
			name: "Version passes against multiple passing requirements",
			fields: fields{
				Major: 9,
				Minor: 0,
			},
			args: args{
				requirements: []string{"8.7+", "9.0-9.0"},
			},
			want: true,
		},
		{
			name:   "Four component version passes against range",
			fields: fields{9, 0, 1, 1},
			args: args{
				requirements: []string{"6.0-10.0"},
			},
			want: true,
		},
		{
			name:   "Four component major-plus-zeroes version passes as bottom of range",
			fields: fields{6, 0, 0, 0},
			args: args{
				requirements: []string{"6.0-10.0"},
			},
			want: true,
		},
		{
			name:   "Version fails because its patch version exceeds range upper boundary",
			fields: fields{10, 0, 1, 0},
			args: args{
				requirements: []string{"6.0-10.0"},
			},
			want: false,
		},
		{
			name: "Two component version fails because it's assumed patch version of zero is less than the patch version of lower boundary",
			fields: fields{
				Major: 6,
				Minor: 0,
			},
			args: args{
				requirements: []string{"6.0.1.0-10.0"},
			},
			want: false,
		},
		{
			name: "Three component version passes four-component-to-infinity version specification",
			fields: fields{
				Major: 2,
				Minor: 0,
				Patch: 0,
			},
			args: args{
				requirements: []string{"1.2.3.4+"},
			},
			want: true,
		},
		{
			name: "Two component version passes three-component specification with plus on lefthand side",
			fields: fields{
				Major: 2,
				Minor: 0,
			},
			args: args{
				requirements: []string{"+1.9.2"},
			},
			want: true,
		},
		{
			name: "Version Major.Minor fails vs lower boundary of three-component range",
			fields: fields{
				Major: 5,
				Minor: 0,
			},
			args: args{
				requirements: []string{"6.0.0.0-10.0.0"},
			},
			want: false,
		},
		{
			name: "Single component version passes single component-plus range",
			fields: fields{
				Major: 5,
			},
			args: args{
				requirements: []string{"4+"},
			},
			want: true,
		},
		{
			name:   "Four component version passes zero to two-component range",
			fields: fields{3, 1, 2, 3},
			args: args{
				requirements: []string{"0-4.1"},
			},
			want: true,
		},
		{
			name: "Four component version fails outside of upper boundary",
			fields: fields{
				Major: 11,
				Minor: 1,
				Patch: 2,
			},
			args: args{
				requirements: []string{"6-10.0"},
			},
			want: false,
		},
		{
			name: "Two component version passes four component range",
			fields: fields{
				Major: 4,
				Minor: 0,
			},
			args: args{
				requirements: []string{"4.0.0.0"},
			},
			want: true,
		},
		{
			name:   "Four component version passes vs two single-component, single element ranges",
			fields: fields{4, 0, 0, 0},
			args: args{
				requirements: []string{"5", "4"},
			},
			want: true,
		},
		{
			name:   "Four component version fails when it exceeds upper boundary of two-component to two-component range",
			fields: fields{6, 0, 0, 0},
			args: args{

				requirements: []string{"4.1-5.7"},
			},
			want: false,
		},
		{
			name:   "Four-component version fails vs two single-component, single element ranges",
			fields: fields{4, 1, 0, 0},
			args: args{
				requirements: []string{"5", "4"},
			},
			want: false,
		},
		{
			name: "Single-component version passes vs single-component zero-to-infinity range",
			fields: fields{
				Major: 10,
			},
			args: args{
				requirements: []string{"0+"},
			},
			want: true,
		},
		{
			name: "Single-component version passes vs two-component zero-to-infinity range",
			fields: fields{
				Major: 10,
			},
			args: args{
				requirements: []string{"0.0+"},
			},
			want: true,
		},
		{
			name: "Single-component version passes vs three-component zero-to-infinity range",
			fields: fields{
				Major: 10,
			},
			args: args{
				requirements: []string{"0.0.0+"},
			},
			want: true,
		},
		{
			name: "Single-component version passes vs four-component zero-to-infinity range",
			fields: fields{
				Major: 10,
			},
			args: args{
				requirements: []string{"0.0.0.0+"},
			},
			want: true,
		},
		{
			name: "Two-component version passes vs single-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
			},
			args: args{
				requirements: []string{"0+"},
			},
			want: true,
		},
		{
			name: "Two-component version passes vs two-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
			},
			args: args{
				requirements: []string{"0.0+"},
			},
			want: true,
		},
		{
			name: "Two-component version passes vs three-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
				Patch: 1,
			},
			args: args{
				requirements: []string{"0.0.0+"},
			},
			want: true,
		},
		{
			name: "Two-component version passes vs four-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
				Patch: 1,
			},
			args: args{
				requirements: []string{"0.0.0.0+"},
			},
			want: true,
		},
		{
			name: "Three-component version passes vs single-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
				Patch: 1,
			},
			args: args{
				requirements: []string{"0+"},
			},
			want: true,
		},
		{
			name: "Three-component version passes vs two-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
				Patch: 1,
			},
			args: args{
				requirements: []string{"0.0+"},
			},
			want: true,
		},
		{
			name: "Three-component version passes vs three-component zero-to-infinity range",
			fields: fields{
				Major: 10,
				Minor: 0,
				Patch: 1,
			},
			args: args{
				requirements: []string{"0.0.0+"},
			},
			want: true,
		},
		{
			name:   "Four-component version passes vs single-component zero-to-infinity range",
			fields: fields{10, 0, 1, 2},
			args: args{
				requirements: []string{"0+"},
			},
			want: true,
		},
		{
			name:   "Four-component version passes vs two-component zero-to-infinity range",
			fields: fields{10, 0, 1, 2},
			args: args{
				requirements: []string{"0.0+"},
			},
			want: true,
		},
		{
			name:   "Four-component version passes vs three-component zero-to-infinity range",
			fields: fields{10, 0, 1, 2},
			args: args{
				requirements: []string{"0.0.0+"},
			},
			want: true,
		},
		{
			name:   "Four-component version passes vs four-component zero-to-infinity range",
			fields: fields{10, 0, 1, 2},
			args: args{
				requirements: []string{"0.0.0.0+"},
			},
			want: true,
		},
		{
			name:   "Invalid three-part version range generates an error",
			fields: fields{10, 0, 1, 0},
			args: args{
				requirements: []string{"4-7-9"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "Invalid requirements input fails",
			fields: fields{10, 0, 1, 0},
			args: args{
				requirements: []string{"7V3^7mpR@Fg9MH^rZu@NXh"},
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "Call on empty receiver fails",
			fields: fields{},
			args: args{
				requirements: []string{"4-7.2"},
			},
			want: false,
		},
		{
			name:   "Empty requirements input fails with error",
			fields: fields{10, 0, 1, 0},
			args: args{
				requirements: []string{""},
			},
			want:    false,
			wantErr: true,
		},
		{
			name:   "Range containing spaces passes",
			fields: fields{7, 0, 1, 0},
			args: args{
				requirements: []string{"7.0 - 9.2"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Ver{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
				Build: tt.fields.Build,
			}
			got, err := v.CheckCompatibility(tt.args.requirements)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ver.CheckCompatibility() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ver.CheckCompatibility() = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ = Describe("version parser helper functions", func() {

	Describe("ver.String()", func() {
		var (
			version Ver
			result  string
		)
		JustBeforeEach(func() {
			result = version.String()
		})
		Context("When called on Ver with non-default values", func() {
			BeforeEach(func() {
				version = Ver{
					Major: 1,
					Minor: 2,
					Patch: 3,
					Build: 4,
				}
			})
			It("Should return expected string", func() {
				Expect(result).To(Equal("1.2.3.4"))
			})
		})

		Context("When called on Ver with default values", func() {
			BeforeEach(func() {
				version = Ver{}
			})
			It("Should return expected string", func() {
				Expect(result).To(Equal("0.0.0.0"))
			})
		})

	})
})
