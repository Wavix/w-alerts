package utils

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"regexp"

	"github.com/gofrs/uuid"
)

// GetValueFromMap returns the value of nested key in a map (ex: a.b.c.d)
func GetValueFromMap(m map[string]interface{}, keyPath string) (interface{}, bool) {
	if keyPath == "" {
		keyPath = "value"
	}

	key := keyPath
	keys := strings.Split(key, ".")

	var value interface{}
	value = m

	for _, key := range keys {
		castedVal, ok := value.(map[string]interface{})
		if !ok {
			return nil, false
		}

		value, ok = castedVal[key]
		if !ok {
			return nil, false
		}
	}

	return value, true
}

func ToNumber(value interface{}) float64 {
	var result float64

	switch v := value.(type) {
	case int:
		result = float64(v)
	case float64:
		result = v
	default:
		return 0
	}

	return result
}

func IsString(value interface{}) bool {
	if _, ok := value.(string); ok {
		return true
	}
	return false
}

func GenerateRuleUUID(fileName string, ruleName string) string {
	hasher := sha1.New()
	hasher.Write([]byte(fileName))
	hasher.Write([]byte(ruleName))
	hashedBytes := hasher.Sum(nil)

	uuid, err := uuid.FromBytes(hashedBytes[:16])
	if err != nil {
		panic(err)
	}
	return uuid.String()
}

func ReplacePlaceholders(template string, values []interface{}, responseData map[string]interface{}) string {
	var result strings.Builder

	if !strings.Contains(template, "{}") {
		return template
	}

	parts := strings.Split(template, "{}")

	for i, part := range parts {
		result.WriteString(part)

		if i < len(values) {
			result.WriteString(fmt.Sprintf("%v", values[i]))
		} else if i < len(parts)-1 {
			result.WriteString("0")
		}
	}

	re := regexp.MustCompile(`\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(result.String(), -1)
	for _, match := range matches {
		placeholder := match[0]
		keyPath := match[1]
		if value, ok := GetValueFromMap(responseData, keyPath); ok {
			resultStr := strings.Replace(result.String(), placeholder, fmt.Sprintf("%v", value), 1)
			result.Reset()
			result.WriteString(resultStr)
		}
	}
	return result.String()
}
