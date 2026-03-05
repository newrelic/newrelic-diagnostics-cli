package obfuscate

import "testing"

func TestObfuscateSensitiveValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "Empty string returns empty",
			value: "",
			want:  "",
		},
		{
			name:  "Short string (4 chars) returns all asterisks",
			value: "abcd",
			want:  "****",
		},
		{
			name:  "Exactly 6 chars returns all asterisks",
			value: "abcdef",
			want:  "******",
		},
		{
			name:  "7 chars shows first 6 + 1 asterisk",
			value: "abcdefg",
			want:  "abcdef*",
		},
		{
			name:  "Standard 40-char license key",
			value: "0123456789abcdef0123456789abcdef01234567",
			want:  "012345**********************************",
		},
		{
			name:  "32-char API key",
			value: "THIS-ISNTAREALAPIKEYOKAYTHANKYOU",
			want:  "THIS-I**************************",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ObfuscateSensitiveValue(tt.value); got != tt.want {
				t.Errorf("ObfuscateSensitiveValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSensitiveKey(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		format string
		want   bool
	}{
		{
			name:   "YAML license_key exact match",
			key:    "license_key",
			format: "yaml",
			want:   true,
		},
		{
			name:   "YAML LICENSE_KEY uppercase",
			key:    "LICENSE_KEY",
			format: "yaml",
			want:   true,
		},
		{
			name:   "YAML non-sensitive key",
			key:    "app_name",
			format: "yaml",
			want:   false,
		},
		{
			name:   "JSON licenseKey",
			key:    "licenseKey",
			format: "json",
			want:   true,
		},
		{
			name:   "XML password",
			key:    "password",
			format: "xml",
			want:   true,
		},
		{
			name:   "INI newrelic.license",
			key:    "newrelic.license",
			format: "ini",
			want:   true,
		},
		{
			name:   "Unknown format",
			key:    "license_key",
			format: "unknown",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSensitiveKey(tt.key, tt.format); got != tt.want {
				t.Errorf("IsSensitiveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
