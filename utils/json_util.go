package utils

import (
	"encoding/json"
	"net/http"
)

func GetJsonData(data any) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		panic("Can't parse data to JSON. This should never happen!")
	}
	return jsonData
}

func WriteResponseStatusCode(data any, statusCode int, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)
	jsonData := GetJsonData(data)
	writer.Write(jsonData)
}

func WriteJsonResponse(data any, writer http.ResponseWriter) {
	writer.Header().Set("Content-Type", "application/json")
	jsonData := GetJsonData(data)
	writer.Write(jsonData)
}