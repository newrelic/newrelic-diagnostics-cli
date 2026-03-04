package obfuscate

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var LicenseKeyPatterns = map[string][]string{
	"yaml": {"license_key", "licenseKey"},
	"json": {"licenseKey", "license_key"},
	"xml":  {"licenseKey"},
	"ini":  {"newrelic.license", "license_key"},
}

var APIKeyPatterns = map[string][]string{
	"yaml": {"api_key", "apiKey"},
	"json": {"api_key", "apiKey"},
	"xml":  {"apiKey"},
	"ini":  {"api_key"},
}

var SensitiveKeyPatterns = map[string][]string{
	"yaml": {
		"license_key", "licenseKey", "proxy_pass", "proxy_password", "proxy_user",
		"password", "pass", "secret", "token", "api_key", "apiKey", "jmx_pass", "jmx_user",
	},
	"json": {
		"licenseKey", "license_key", "api_key", "apiKey", "proxy_pass",
		"password", "secret", "token",
	},
	"xml": {
		"licenseKey", "password", "secret", "apiKey",
	},
	"ini": {
		"newrelic.license", "license_key", "proxy_pass", "password", "api_key",
	},
}

func ObfuscateSensitiveValue(value string) string {
	if value == "" {
		return ""
	}
	keyLen := len(value)
	if keyLen <= 6 {
		return strings.Repeat("*", keyLen)
	}
	return value[:6] + strings.Repeat("*", keyLen-6)
}

func FullyObfuscate(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("*", len(value))
}

func ObfuscateByKeyType(key, value, format string) string {
	if isLicenseOrAPIKey(key, format) {
		return ObfuscateSensitiveValue(value)
	}
	return FullyObfuscate(value)
}

func isLicenseOrAPIKey(key, format string) bool {
	keyLower := strings.ToLower(key)

	if patterns, ok := LicenseKeyPatterns[format]; ok {
		for _, pattern := range patterns {
			if strings.EqualFold(pattern, keyLower) {
				return true
			}
		}
	}

	if patterns, ok := APIKeyPatterns[format]; ok {
		for _, pattern := range patterns {
			if strings.EqualFold(pattern, keyLower) {
				return true
			}
		}
	}

	return false
}

func IsSensitiveKey(key string, format string) bool {
	patterns, ok := SensitiveKeyPatterns[format]
	if !ok {
		return false
	}

	keyLower := strings.ToLower(key)
	for _, pattern := range patterns {
		if strings.EqualFold(pattern, keyLower) {
			return true
		}
	}
	return false
}

var envLicenseKeyPattern = regexp.MustCompile("(?i)(license|api).*key")

type ObfuscatedMap struct {
	data          map[string][]string
	obfuscateKeys bool
}

func NewObfuscatedMap(data map[string][]string, obfuscateKeys bool) ObfuscatedMap {
	return ObfuscatedMap{data: data, obfuscateKeys: obfuscateKeys}
}

func (o ObfuscatedMap) MarshalJSON() ([]byte, error) {
	obfuscated := make(map[string][]string)

	if o.obfuscateKeys {
		for k, v := range o.data {
			obfuscated[ObfuscateSensitiveValue(k)] = v
		}
	} else {
		for k, v := range o.data {
			if len(v) > 0 && envLicenseKeyPattern.MatchString(k) {
				obfuscated[k] = []string{ObfuscateSensitiveValue(v[0])}
			} else {
				obfuscated[k] = v
			}
		}
	}

	return json.Marshal(obfuscated)
}

func WriteObfuscatedConfigFileToZip(path string, storeName string, dst *zip.Writer) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	obfuscatedContent, err := ObfuscateConfigContent(path, content)
	if err != nil {
		return fmt.Errorf("failed to obfuscate config file: %w", err)
	}

	header, err := zip.FileInfoHeader(stat)
	if err != nil {
		return fmt.Errorf("failed to create zip header: %w", err)
	}

	header.Name = filepath.ToSlash("nrdiag-output/" + storeName)
	header.Method = zip.Deflate

	writer, err := dst.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create zip entry: %w", err)
	}

	_, err = writer.Write(obfuscatedContent)
	if err != nil {
		return fmt.Errorf("failed to write obfuscated content: %w", err)
	}

	return nil
}
