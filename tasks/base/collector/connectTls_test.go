package collector

import (
	"errors"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestBaseCollectorConnectTls_prepareCollectorErrorResult(t *testing.T) {
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
			p := BaseCollectorTLS{
				upstream: tt.fields.upstream,
			}
			got := p.prepareCollectorErrorResult(tt.args.e)

			if got.Status != tt.want {
				t.Errorf("prepareCollectorErrorResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectTls_prepareResponseErrorResult(t *testing.T) {
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
			p := BaseCollectorTLS{
				upstream: tt.fields.upstream,
			}
			got := p.prepareResponseErrorResult(tt.args.e, tt.args.statusCode)
			if got.Status != tt.want {
				t.Errorf("prepareResponseErrorResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectTls_prepareResult(t *testing.T) {
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
				body:       "<tr><th>TLS Version</th><td>TLSv1.2</td></tr>",
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
			p := BaseCollectorTLS{
				upstream: tt.fields.upstream,
			}
			got := p.prepareResult(tt.args.body, tt.args.statusCode)
			if got.Status != tt.want {
				t.Errorf("prepareResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorConnectTls_Execute(t *testing.T) {
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
			name: "should return warning result on non-200 status code",
			fields: fields{
				httpGetter: mockUnsuccessfulRequest400,
			},
			args: args{
				op: tasks.Options{},
			},
			want: tasks.Warning,
		},
		{
			name: "should return failed result on failed request",
			fields: fields{
				httpGetter: mockUnsuccessfulRequestError,
			},
			args: args{
				op: tasks.Options{},
			},
			want: tasks.Failure,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			p := BaseCollectorTLS{
				upstream:   tt.fields.upstream,
				httpGetter: tt.fields.httpGetter,
			}
			got := p.Execute(tt.args.op, tt.args.upstream)
			if got.Status != tt.want {
				t.Errorf("BaseCollectorTLS.Execute() = %v, want %v", got.Status, tt.want)
			}
		})
	}
}

func TestBaseCollectorTLS_checkTlsVerMeetsReqs(t *testing.T) {
	type args struct {
		ver string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should succeed if tls 1.2 is detected",
			args: args{
				ver: "TLSv1.2",
			},
			want: true,
		},
		{
			name: "should succeed if tls 1.3 is detected",
			args: args{
				ver: "TLSv1.3",
			},
			want: true,
		},
		{
			name: "should fail if tls 1.1 is detected",
			args: args{
				ver: "TLSv1.1",
			},
			want: false,
		},
		{
			name: "should fail if tls 1.0 is detected",
			args: args{
				ver: "TLSv1.0",
			},
			want: false,
		},
		{
			name: "should fail if non-tls is detected",
			args: args{
				ver: "SSLv3.0",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorTLS{}
			if got := p.checkTlsVerMeetsReqs(tt.args.ver); got != tt.want {
				t.Errorf("BaseCollectorTLS.checkTlsVerMeetsReqs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseCollectorTLS_parseTlsStringFromHtml(t *testing.T) {
	type args struct {
		body string
	}
	goodHtmlTls12 := "<html><head><title>New Relic Connection Test</title></head><body><h3>Detected Connection Settings</h3><table border='1' cellpadding='10' style='border-collapse: collapse; padding'><tr><th>Client IP</th><td>204.195.113.246</td></tr><tr><th>TLS Version</th><td>TLSv1.2</td></tr><tr><th>Cipher</th><td>ECDHE-RSA-AES128-GCM-SHA256</td></tr></table></body></html>"
	goodHtmlNoTls := "<html><head><title>New Relic Connection Test</title></head><body><h3>Detected Connection Settings</h3><table border='1' cellpadding='10' style='border-collapse: collapse; padding'><tr><th>Client IP</th><td>204.195.113.246</td></tr><tr></tr><tr><th>Cipher</th><td>ECDHE-RSA-AES128-GCM-SHA256</td></tr></table></body></html>"
	badHtml := "not html"

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "should find the tls version",
			args:    args{body: goodHtmlTls12},
			want:    "TLSv1.2",
			wantErr: false,
		},
		{
			name:    "should fail if it can't find tls version",
			args:    args{body: goodHtmlNoTls},
			want:    "",
			wantErr: true,
		},
		{
			name:    "should fail if html is malformed",
			args:    args{body: badHtml},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := BaseCollectorTLS{}
			got, err := p.parseTlsStringFromHtml(tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseCollectorTLS.parseTlsStringFromHtml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("BaseCollectorTLS.parseTlsStringFromHtml() = %v, want %v", got, tt.want)
			}
		})
	}
}
