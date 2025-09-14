package jsonUtils

import "encoding/json"

func MapToJson(data map[string]string) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func JsonToMap(jsonData string) map[string]string {
	var data map[string]string
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil
	}
	return data
}
