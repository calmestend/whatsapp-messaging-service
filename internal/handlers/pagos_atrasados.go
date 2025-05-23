package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
)

// @TODO: Create utils function to handle payloadData validations
// @TODO: Change PagosAtrasadosRequest.Amount to int and hardcode "$"

// Struct with parameters for template "pagos_atrasados" message
type PagosAtrasadosRequest struct {
	models.MessageRequest
	Item        string `json:"item" validate:"required"`
	Amount      string `json:"amount" validate:"required"`
	Days        int    `json:"days" validate:"required,min=1"`
	Payment_url string `json:"payment_url" validate:"required,http_url"`
	Support_url string `json:"support_url" validate:"required,http_url"`
}

// Handle template "pagos_atrasados" from the Whatsapp Business API from Meta
func PagosAtrasados(w http.ResponseWriter, r *http.Request) {
	var payloadData PagosAtrasadosRequest
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
