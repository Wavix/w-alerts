package utils

import (
	"encoding/json"
	"fmt"
)

func JSONToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return result, nil
}

func JsonFormat(jsonStr string) string {
	var jsonData map[string]interface{}

	err := json.Unmarshal([]byte(jsonStr), &jsonData)
	if err != nil {
		return ""
	}

	compactJSON, err := json.Marshal(jsonData)
	if err != nil {
		return ""
	}

	return string(compactJSON)
}
