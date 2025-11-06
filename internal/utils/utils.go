package utils

import "encoding/json"

func ParseJSONBody(body []byte) map[string]any {
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}
	return parsed
}
