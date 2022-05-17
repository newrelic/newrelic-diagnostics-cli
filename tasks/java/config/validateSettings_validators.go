package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type ValidationStatus string

const (
	Valid   ValidationStatus = "Valid"
	Invalid ValidationStatus = "Invalid"
	Unknown ValidationStatus = "Unknown"
)

type ValidationResult struct {
	Kind    string
	Key     string
	Value   interface{}
	Status  ValidationStatus
	Message string
}

func ValidateSetting(value interface{}, kind string) ValidationResult {
	validator, exists := ValidatorForType[kind]
	if !exists {
		validator = ValidateUnknown
	}
	result := validator(value)
	if result.Status == "" {
		result.Status = Valid
	} else {
		result.Value = value
	}
	return result
}

var ValidatorForType = map[string]SettingValidator{
	"Integer":              ValidateInteger,
	"Float":                ValidateFloat,
	"AppName":              ValidateAppName,
	"Boolean":              ValidateBoolean,
	"StatusCodeList":       ValidateStatusCodeList,
	"LabelList":            ValidateLabelList,
	"String":               ValidateString,
	"ProxyScheme":          ValidateProxyScheme,
	"LogLevel":             ValidateLogLevel,
	"RecordSql":            ValidateRecordSql,
	"TransactionThreshold": ValidateTransactionThreshold,
}

type SettingValidator func(value interface{}) ValidationResult

func ValidateInteger(value interface{}) (result ValidationResult) {
	if value == nil {
		return
	}
	if _, isInt := value.(int); isInt {
		return
	}
	if strValue, isStr := value.(string); isStr {
		if _, err := strconv.ParseInt(strValue, 10, 64); err == nil {
			return
		}
	}
	result.Status = Invalid
	result.Message = fmt.Sprintf("invalid integer value (%v)", value)
	return
}

func ValidateFloat(value interface{}) (result ValidationResult) {
	if value == nil {
		return
	}
	switch v := value.(type) {
	case float32, float64, int: // fine
		return
	case string:
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			return
		}
	}
	result.Status = Invalid
	result.Message = fmt.Sprintf("invalid value for float (%v)", value)
	return
}

func ValidateAppName(value interface{}) (result ValidationResult) {
	if value == nil {
		result.Status = Invalid
		result.Message = "must not be empty"
	} else if strValue, isStr := value.(string); !isStr {
		result.Status = Invalid
		result.Message = fmt.Sprintf("should be a string (not %T)", value)
	} else {
		if strValue == "" {
			result.Status = Invalid
			result.Message = "must not be empty"
		}
		n := strings.Count(strValue, ";")
		if n > 2 {
			result.Status = Invalid
			result.Message = fmt.Sprintf("at most three semicolon-separated values are allowed - got %v", n)
		}
	}
	return
}

func ValidateBoolean(value interface{}) (result ValidationResult) {
	if value == nil {
		return
	}
	if _, isBool := value.(bool); !isBool {
		theString := strings.ToLower(value.(string))
		if theString != "true" && theString != "false" {
			result.Status = Invalid
			result.Message = "boolean values must be \"true\" or \"false\" (case-insensitive) only"
		}
	}
	return
}

var invalidStatusCode = regexp.MustCompile("[^0-9-,]")

func ValidateStatusCodeList(value interface{}) (result ValidationResult) {
	if value == nil {
		return
	}
	switch v := value.(type) {
	case string:
		if invalidStatusCode.MatchString(v) {
			result.Status = Invalid
			result.Message = "should be a comma-separated list of status codes or status code ranges"
		} else {
			for _, scRange := range strings.Split(v, ",") {
				var last int = -1
				for j, sc := range strings.Split(scRange, "-") {
					if j > 1 {
						result.Status = Invalid
						result.Message = "contain only two values"
					}
					scInt, err := strconv.Atoi(sc)
					if err != nil {
						result.Status = Invalid
						result.Message = "must consist only of numbers"
					} else if scInt < 0 || scInt > 1000 {
						result.Status = Invalid
						result.Message = "must be within the range 0-1000 inclusive"
					} else if scInt <= last {
						result.Status = Invalid
						result.Message = "left side of range must be less than right side"
					}
					last = scInt
				}
			}
		}
	case int:
		if v < 0 || v > 1000 {
			result.Status = Invalid
			result.Message = "must be within the range 0-1000 inclusive"
		}
	default:
		result.Status = Invalid
		result.Message = "should be a comma-separated list of status codes or status code ranges"
	}

	return
}

func ValidateLabelList(value interface{}) (result ValidationResult) {
	message := "must be one or more key:value pairs, separated by semicolons, or YAML key: value"
	if value == nil {
		return
	}
	if yaml, ok := value.(map[interface{}]interface{}); ok {
		for _, v := range yaml {
			if str, ok := v.(string); ok && str == "" {
				result.Status = Invalid
				result.Message = fmt.Sprintf("empty label %s", str)
			}
		}
		return
	}

	if str, ok := value.(string); ok {
		pairs := strings.Split(str, ";")
		for _, pair := range pairs {
			keyValue := strings.Split(pair, ":")
			if len(keyValue) != 2 || keyValue[0] == "" || keyValue[1] == "" {
				result.Status = Invalid
				result.Message = message
			}
		}
	} else {
		result.Status = Invalid
		result.Message = message
	}
	return
}

func ValidateString(value interface{}) (result ValidationResult) {
	// FIXME: is there anything to actually check for here?
	return
}

func ValidateProxyScheme(value interface{}) (result ValidationResult) {
	return ValidateEnum(value, []string{"http", "https"})
}

func ValidateLogLevel(value interface{}) (result ValidationResult) {
	return ValidateEnum(value, []string{"off", "severe", "warning", "info", "fine", "finer", "finest"})

}

func ValidateRecordSql(value interface{}) (result ValidationResult) {
	return ValidateEnum(value, []string{"off", "raw", "obfuscated"})
}

func ValidateTransactionThreshold(value interface{}) (result ValidationResult) {
	if value == nil {
		return
	}
	if _, ok := value.(float64); ok {
		return
	}
	if _, ok := value.(int); ok {
		return
	}
	if str, ok := value.(string); ok {
		if str == "apdex_f" {
			return
		}
	}
	result.Status = Invalid
	result.Message = "must be a float or \"apdex_f\""
	return
}

func ValidateEnum(value interface{}, enumValues []string) (result ValidationResult) {
	if value == nil {
		return
	}
	if stringValue, ok := value.(string); ok {
		for _, enumValue := range enumValues {
			if strings.EqualFold(stringValue, enumValue) {
				return
			}
		}
	}
	result.Status = Invalid
	result.Message = "value must be one of: " + strings.Join(enumValues, ",")
	return
}

func ValidateUnknown(value interface{}) (result ValidationResult) {
	result.Status = "Unknown"
	return
}
