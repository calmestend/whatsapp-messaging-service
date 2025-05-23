package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
)

// @TODO: Create utils function to handle payloadData validations

// Struct with parameters for template "encuesta_nps" message
type EncuestaNpsRequest struct {
	models.MessageRequest
	Address string `json:"address" validate:"required"`
	Date    string `json:"date" validate:"required"`
}

// Handle template "encuesta_nps" from the Whatsapp Business API from Meta
func EncuestaNps(w http.ResponseWriter, r *http.Request) {
	var payloadData EncuestaNpsRequest
	err := json.NewDecoder(r.Body).Decode(&payloadData)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	validate := validator.New()
	err = validate.Struct(payloadData)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		http.Error(w, fmt.Sprintf("Validation error: %s", errors), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	fmt.Fprintln(w, "Message sent")
	json.NewEncoder(w).Encode(payloadData)
}
