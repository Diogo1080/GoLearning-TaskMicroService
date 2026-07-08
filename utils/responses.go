package utils

import (
	"backendGo/models"
	"encoding/json"
	"net/http"
)

func SendErrorResponse(res http.ResponseWriter, errorMessage string, statusCode int) {
	resp := models.ErrorResponse{Error: errorMessage}
	respBytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(res, "Error:", http.StatusInternalServerError)
		return
	}
	http.Error(res, string(respBytes), statusCode)
}
