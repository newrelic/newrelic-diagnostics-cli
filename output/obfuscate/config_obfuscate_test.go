package obfuscate

import (
	"strings"
	"testing"
)

func Test_obfuscateYAML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name: "YAML with license_key",
			content: `license_key: abc123def456ghi789
app_name: MyApp`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name: "YAML with proxy_pass",
			content: `proxy_host: proxy.example.com
proxy_pass: secretpassword123`,
			want:    "*****************",
			wantErr: false,
		},
		{
			name: "YAML nested structure",
			content: `common: &default
  license_key: mykey123456789
production:
  <<: *default`,
			want:    "mykey1",
			wantErr: false,
		},
		{
			name:    "Invalid YAML",
			content: `invalid: yaml: content:`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := obfuscateYAML([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("obfuscateYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(got), tt.want) {
				t.Errorf("obfuscateYAML() result should contain %v, got %v", tt.want, string(got))
			}
		})
	}
}

func Test_obfuscateJSON(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "JSON with licenseKey",
			content: `{"licenseKey": "abc123def456ghi789", "appName": "MyApp"}`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "JSON with api_key",
			content: `{"api_key": "secret123456", "name": "test"}`,
			want:    "secret",
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			content: `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := obfuscateJSON([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("obfuscateJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(got), tt.want) {
				t.Errorf("obfuscateJSON() result should contain %v, got %v", tt.want, string(got))
			}
		})
	}
}

func Test_obfuscateXML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "XML with licenseKey element",
			content: `<configuration><licenseKey>abc123def456ghi789</licenseKey></configuration>`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "XML with licenseKey attribute",
			content: `<service licenseKey="abc123def456ghi789" />`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "XML with password",
			content: `<database><password>secretpass123</password></database>`,
			want:    "*************",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := obfuscateXML([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("obfuscateXML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(got), tt.want) {
				t.Errorf("obfuscateXML() result should contain %v, got %v", tt.want, string(got))
			}
		})
	}
}

func Test_obfuscateINI(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr bool
	}{
		{
			name: "INI with license_key",
			content: `app_name = MyApp
license_key = abc123def456ghi789`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name: "INI with newrelic.license",
			content: `newrelic.license=abc123def456ghi789
newrelic.appname=MyApp`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name: "INI with comments",
			content: `; This is a comment
license_key = abc123def456ghi789
# Another comment
app_name = test`,
			want:    "abc123",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := obfuscateINI([]byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("obfuscateINI() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(got), tt.want) {
				t.Errorf("obfuscateINI() result should contain %v, got %v", tt.want, string(got))
			}
		})
	}
}

func TestObfuscateConfigContent(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    string
		wantErr bool
	}{
		{
			name:    "YAML file",
			path:    "/path/to/newrelic.yml",
			content: `license_key: abc123def456`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "JSON file",
			path:    "/path/to/config.json",
			content: `{"licenseKey": "abc123def456"}`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "XML file",
			path:    "/path/to/newrelic.config",
			content: `<licenseKey>abc123def456</licenseKey>`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "INI file",
			path:    "/path/to/newrelic.ini",
			content: `license_key=abc123def456`,
			want:    "abc123",
			wantErr: false,
		},
		{
			name:    "Unsupported extension",
			path:    "/path/to/file.txt",
			content: `some content`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ObfuscateConfigContent(tt.path, []byte(tt.content))
			if (err != nil) != tt.wantErr {
				t.Errorf("ObfuscateConfigContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !strings.Contains(string(got), tt.want) {
				t.Errorf("ObfuscateConfigContent() result should contain %v, got %v", tt.want, string(got))
			}
		})
	}
}
