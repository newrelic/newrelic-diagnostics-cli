//go:build windows
// +build windows

package profiler

import (
	"reflect"
	"testing"
)

func Test_validateTLSRegKeys(t *testing.T) {
	tests := []struct {
		name    string
		want    *TLSRegKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "First test",
			want:    &TLSRegKey{0, 1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateTLSRegKeys()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTLSRegKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateTLSRegKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
