package jsontools

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FlattenJSON converts nested JSON to flat key-value pairs
// Example: {"user": {"name": "John"}} → {"user.name": "John"}
func FlattenJSON(data map[string]interface{}, prefix string) map[string]string {
	result := make(map[string]string)
	flattenHelper(data, prefix, result)
	return result
}

func flattenHelper(data map[string]interface{}, prefix string, result map[string]string) {
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested objects
			flattenHelper(v, fullKey, result)
		case []interface{}:
			// Convert arrays to JSON string
			jsonBytes, _ := json.Marshal(v)
			result[fullKey] = string(jsonBytes)
		default:
			// Store primitive values as strings
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}
}

// UnflattenJSON rebuilds nested JSON from flat structure
func UnflattenJSON(flat map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range flat {
		parts := strings.Split(key, ".")
		current := result

		for i, part := range parts {
			if i == len(parts)-1 {
				// Last part - set the value
				// Try to parse arrays back
				if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
					var arrValue interface{}
					if err := json.Unmarshal([]byte(value), &arrValue); err == nil {
						current[part] = arrValue
					} else {
						current[part] = value
					}
				} else {
					current[part] = value
				}
			} else {
				// Intermediate part - ensure map exists
				if _, exists := current[part]; !exists {
					current[part] = make(map[string]interface{})
				}
				if nested, ok := current[part].(map[string]interface{}); ok {
					current = nested
				}
			}
		}
	}

	return result
}

// ParseJSON parses JSON bytes into a map
func ParseJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return result, nil
}

// ValidateTranslationFile ensures JSON structure is valid
func ValidateTranslationFile(content []byte) error {
	var data map[string]interface{}
	if err := json.Unmarshal(content, &data); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("empty JSON object")
	}
	return nil
}
