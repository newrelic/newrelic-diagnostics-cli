package collector

import (
	"errors"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestBaseCollectorConnectUS_prepareEarlyResult(t *testing.T) {

	payloadregionDetectUSEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{"us01", "eu01"}},
	}
	payloadregionDetectEmpty := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{}},
	}
	payloadregionDetectUS := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{"us01"}},
	}
	payloadregionDetectEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{"eu01"}},
	}

	type fields struct {
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name   string
		fields fields
		want   tasks.Result
	}{
		{
			name:   "should return an empty result if no regions detected",
			fields: fields{upstream: payloadregionDetectEmpty},
			want:   tasks.Result{},
		},
		{
			name:   "should return an empty result if US Region is only region detected",
			fields: fields{upstream: payloadregionDetectUS},
			want:   tasks.Result{},
		},
		{
			name:   "should return an empty result if US Region is among regions detected",
			fields: fields{upstream: payloadregionDetectUSEU},
			want:   tasks.Result{},
		},
		{
			name:   "should return a None result if US Region not detected among other regions",
			fields: fields{upstream: payloadregionDetectEU},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "US Region not detected, skipping US collector connect check",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorConnectUS{
				upstream: tt.fields.upstream,
			}
			if got := p.prepareEarlyResult(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseCollectorConnectUS.prepareEarlyResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectUS_prepareCollectorErrorResult(t *testing.T) {
	sampleError := errors.New("received HTTP Error: this is an error")
	type fields struct {
		upstream map[string]tasks.Result
	}
	type args struct {
		e error
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   tasks.Status
	}{
		{
			name:   "should return a result with Status 'Failure' when given an error",
			args:   args{e: sampleError},
			fields: fields{},
			want:   tasks.Failure,
		},
		{
			name:   "should produce an empty result when given a nil error",
			args:   args{},
			fields: fields{},
			want:   tasks.Result{}.Status,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorConnectUS{
				upstream: tt.fields.upstream,
			}
			got := p.prepareCollectorErrorResult(tt.args.e)

			if got.Status != tt.want {
				t.Errorf("prepareCollectorErrorResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectUS_prepareResponseErrorResult(t *testing.T) {
	sampleError := errors.New("could not parse response body")
	type fields struct {
		upstream map[string]tasks.Result
	}
	type args struct {
		e          error
		statusCode string
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   tasks.Status
	}{
		{
			name: "should return a Warning result if error is not nil",
			args: args{
				e:          sampleError,
				statusCode: "200",
			},
			fields: fields{},
			want:   tasks.Warning,
		},
		{
			name: "should return an empty result if error is nil",
			args: args{
				statusCode: "200",
			},
			fields: fields{},
			want:   tasks.Result{}.Status,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorConnectUS{
				upstream: tt.fields.upstream,
			}
			got := p.prepareResponseErrorResult(tt.args.e, tt.args.statusCode)
			if got.Status != tt.want {
				t.Errorf("prepareResponseErrorResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectUS_prepareResult(t *testing.T) {
	type fields struct {
		upstream map[string]tasks.Result
	}
	type args struct {
		body       string
		statusCode string
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   tasks.Status
	}{
		{
			name: "should return a Success result given '200' status code",
			args: args{
				body:       "{}",
				statusCode: "404",
			},
			fields: fields{},
			want:   tasks.Success,
		},
		{
			name: "should return a Warning result given a non '200' status code",
			args: args{
				body:       "Document has moved permanently",
				statusCode: "301",
			},
			fields: fields{},
			want:   tasks.Warning,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorConnectUS{
				upstream: tt.fields.upstream,
			}
			got := p.prepareResult(tt.args.body, tt.args.statusCode)
			if got.Status != tt.want {
				t.Errorf("prepareResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectUS_Execute(t *testing.T) {

	payloadregionDetectUSEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{"us01", "eu01"}},
	}

	payloadregionDetectEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{"eu01"}},
	}

	payloadregionDetectEmpty := map[string]tasks.Result{
		"Base/Config/RegionDetect": {Payload: []string{}},
	}
	type fields struct {
		upstream   map[string]tasks.Result
		httpGetter requestFunc
	}
	type args struct {
		op       tasks.Options
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   tasks.Status
	}{
		{
			name: "when no regions detected, should return successful result on successful connection",
			fields: fields{
				upstream:   payloadregionDetectEmpty,
				httpGetter: mockSuccessfulRequest200,
			},
			args: args{
				op:       tasks.Options{},
				upstream: payloadregionDetectEmpty,
			},
			want: tasks.Success,
		},
		{
			name: "when US region detected, should return successful result on successful connection",
			fields: fields{
				upstream:   payloadregionDetectUSEU,
				httpGetter: mockSuccessfulRequest200,
			},
			args: args{
				op:       tasks.Options{},
				upstream: payloadregionDetectUSEU,
			},
			want: tasks.Success,
		},
		{
			name: "when US region detected, should return warning result on non-200 status code",
			fields: fields{
				upstream:   payloadregionDetectUSEU,
				httpGetter: mockUnsuccessfulRequest400,
			},
			args: args{
				op:       tasks.Options{},
				upstream: payloadregionDetectUSEU,
			},
			want: tasks.Warning,
		},
		{
			name: "when US region detected, should return failed result on failed request",
			fields: fields{
				upstream:   payloadregionDetectUSEU,
				httpGetter: mockUnsuccessfulRequestError,
			},
			args: args{
				op:       tasks.Options{},
				upstream: payloadregionDetectUSEU,
			},
			want: tasks.Failure,
		},
		{
			name: "when US region not detected, should return none result",
			fields: fields{
				upstream: payloadregionDetectEU,
			},
			args: args{
				op:       tasks.Options{},
				upstream: payloadregionDetectEU,
			},
			want: tasks.None,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := BaseCollectorConnectUS{
				upstream:   tt.fields.upstream,
				httpGetter: tt.fields.httpGetter,
			}
			got := p.Execute(tt.args.op, tt.args.upstream)
			if got.Status != tt.want {
				t.Errorf("BaseCollectorConnectUS.Execute() = %v, want %v", got.Status, tt.want)
			}
		})
	}
}
