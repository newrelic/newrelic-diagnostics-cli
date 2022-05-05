package collector

import (
	"errors"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestBaseCollectorConnectEU_prepareEarlyResult(t *testing.T) {

	payloadregionDetectUSEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{"us01", "eu01"}},
	}
	payloadregionDetectEmpty := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{}},
	}
	payloadregionDetectUS := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{"us01"}},
	}
	payloadregionDetectEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{"eu01"}},
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
			name:   "should return an empty result if EU Region is only region detected",
			fields: fields{upstream: payloadregionDetectEU},
			want:   tasks.Result{},
		},
		{
			name:   "should return an empty result if EU Region is among regions detected",
			fields: fields{upstream: payloadregionDetectUSEU},
			want:   tasks.Result{},
		},
		{
			name:   "should return a None result if EU Region not detected among other regions",
			fields: fields{upstream: payloadregionDetectUS},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "EU Region not detected, skipping EU collector connect check",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorConnectEU{
				upstream: tt.fields.upstream,
			}
			if got := p.prepareEarlyResult(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseCollectorConnectEU.prepareEarlyResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectEU_prepareCollectorErrorResult(t *testing.T) {
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
			p := BaseCollectorConnectEU{
				upstream: tt.fields.upstream,
			}
			got := p.prepareCollectorErrorResult(tt.args.e)

			if got.Status != tt.want {
				t.Errorf("prepareCollectorErrorResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectEU_prepareResponseErrorResult(t *testing.T) {
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
			p := BaseCollectorConnectEU{
				upstream: tt.fields.upstream,
			}
			got := p.prepareResponseErrorResult(tt.args.e, tt.args.statusCode)
			if got.Status != tt.want {
				t.Errorf("prepareResponseErrorResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectEU_prepareResult(t *testing.T) {
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
				body:       "mongrel ==> up (true)",
				statusCode: "200",
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
			p := BaseCollectorConnectEU{
				upstream: tt.fields.upstream,
			}
			got := p.prepareResult(tt.args.body, tt.args.statusCode)
			if got.Status != tt.want {
				t.Errorf("prepareResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectEU_Execute(t *testing.T) {

	payloadregionDetectUSEU := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{"us01", "eu01"}},
	}

	payloadregionDetectUS := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{"us01"}},
	}

	payloadregionDetectEmpty := map[string]tasks.Result{
		"Base/Config/RegionDetect": tasks.Result{Payload: []string{}},
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
			name: "when EU region detected, should return successful result on successful connection",
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
			name: "when EU region detected, should return warning result on non-200 status code",
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
			name: "when EU region detected, should return failed result on failed request",
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
			name: "when EU region not detected detected, should return none result",
			fields: fields{
				upstream: payloadregionDetectUS,
			},
			args: args{
				op:       tasks.Options{},
				upstream: payloadregionDetectUS,
			},
			want: tasks.None,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := BaseCollectorConnectEU{
				upstream:   tt.fields.upstream,
				httpGetter: tt.fields.httpGetter,
			}
			got := p.Execute(tt.args.op, tt.args.upstream)
			if got.Status != tt.want {
				t.Errorf("BaseCollectorConnectEU.Execute() = %v, want %v", got.Status, tt.want)
			}
		})
	}
}
