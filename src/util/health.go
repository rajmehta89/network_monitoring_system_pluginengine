package util

import (
"encoding/json"
)


type HealthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}


func HandleHealthCheck(responseData map[string]interface{}) string {

	response := HealthCheckResponse{
		Status:  "OK",
		Message: "Service is running smoothly",
	}

	responseBytes, err := json.Marshal(response)

	if err != nil {

		return `{"status":"ERROR","message":"Failed to generate response"}`

	}

	return string(responseBytes)

}
