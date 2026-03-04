package obfuscate

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	jsPatterns         = make(map[string]*regexp.Regexp)
	xmlElementPatterns = make(map[string]*regexp.Regexp)
	xmlAttrPatterns    = make(map[string]*regexp.Regexp)
)

func init() {
	for _, key := range SensitiveKeyPatterns["json"] {
		pattern := fmt.Sprintf(`(%s\s*:\s*['"\x60])([^'"\x60]+)(['"\x60])`, regexp.QuoteMeta(key))
		jsPatterns[key] = regexp.MustCompile(pattern)
	}

	for _, key := range SensitiveKeyPatterns["xml"] {
		elementPattern := fmt.Sprintf(`(<%s[^>]*>)([^<]+)(</%s>)`, regexp.QuoteMeta(key), regexp.QuoteMeta(key))
		xmlElementPatterns[key] = regexp.MustCompile(elementPattern)

		attrPattern := fmt.Sprintf(`(%s\s*=\s*")([^"]+)(")`, regexp.QuoteMeta(key))
		xmlAttrPatterns[key] = regexp.MustCompile(attrPattern)
	}
}

func ObfuscateConfigContent(path string, content []byte) ([]byte, error) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".yml", ".yaml":
		return obfuscateYAML(content)
	case ".json":
		return obfuscateJSON(content)
	case ".js":
		return obfuscateJS(content)
	case ".xml", ".config":
		return obfuscateXML(content)
	case ".ini", ".properties", ".cfg":
		return obfuscateINI(content)
	default:
		return content, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func obfuscateYAML(content []byte) ([]byte, error) {
	var data any
	if err := yaml.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	obfuscated := obfuscateMapRecursively(data, "yaml")

	output, err := yaml.Marshal(obfuscated)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return output, nil
}

func obfuscateJSON(content []byte) ([]byte, error) {
	var data any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	obfuscated := obfuscateMapRecursively(data, "json")

	output, err := json.MarshalIndent(obfuscated, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return output, nil
}

func obfuscateJS(content []byte) ([]byte, error) {
	jsContent := string(content)

	for _, key := range SensitiveKeyPatterns["json"] {
		re := jsPatterns[key]
		jsContent = re.ReplaceAllStringFunc(jsContent, func(match string) string {
			parts := re.FindStringSubmatch(match)
			if len(parts) == 4 {
				obfuscated := ObfuscateByKeyType(key, parts[2], "json")
				return parts[1] + obfuscated + parts[3]
			}
			return match
		})
	}

	return []byte(jsContent), nil
}

func obfuscateXML(content []byte) ([]byte, error) {
	result := string(content)

	for _, key := range SensitiveKeyPatterns["xml"] {
		elementRe := xmlElementPatterns[key]
		result = elementRe.ReplaceAllStringFunc(result, func(match string) string {
			parts := elementRe.FindStringSubmatch(match)
			if len(parts) == 4 {
				obfuscated := ObfuscateByKeyType(key, parts[2], "xml")
				return parts[1] + obfuscated + parts[3]
			}
			return match
		})

		attrRe := xmlAttrPatterns[key]
		result = attrRe.ReplaceAllStringFunc(result, func(match string) string {
			parts := attrRe.FindStringSubmatch(match)
			if len(parts) == 4 {
				obfuscated := ObfuscateByKeyType(key, parts[2], "xml")
				return parts[1] + obfuscated + parts[3]
			}
			return match
		})
	}

	return []byte(result), nil
}

func obfuscateINI(content []byte) ([]byte, error) {
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		if IsSensitiveKey(key, "ini") {
			value := strings.TrimSpace(parts[1])
			obfuscated := ObfuscateByKeyType(key, value, "ini")
			lines[i] = parts[0] + "=" + obfuscated
		}
	}

	return []byte(strings.Join(lines, "\n")), nil
}

func obfuscateMapRecursively(data any, format string) any {
	switch v := data.(type) {
	case map[string]any:
		for key, value := range v {
			if IsSensitiveKey(key, format) {
				if strVal, ok := value.(string); ok {
					v[key] = ObfuscateByKeyType(key, strVal, format)
				}
			} else {
				v[key] = obfuscateMapRecursively(value, format)
			}
		}
		return v

	case map[any]any:
		result := make(map[any]any)
		for key, value := range v {
			keyStr, ok := key.(string)
			if ok && IsSensitiveKey(keyStr, format) {
				if strVal, ok := value.(string); ok {
					result[key] = ObfuscateByKeyType(keyStr, strVal, format)
				} else {
					result[key] = value
				}
			} else {
				result[key] = obfuscateMapRecursively(value, format)
			}
		}
		return result

	case []any:
		for i, item := range v {
			v[i] = obfuscateMapRecursively(item, format)
		}
		return v

	default:
		return data
	}
}
