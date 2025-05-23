package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
)

// @TODO: Create utils function to handle payloadData validations
// @TODO: Change EnvioCotizacionRequest.Amount to int and hardcode "$"

// Struct with parameters for template "envio_cotizacion" message
type EnvioCotizacionRequest struct {
	models.MessageRequest
	CustomerName string `json:"customerName" validate:"required"`
	BusinessName string `json:"businessName" validate:"required"`
	Folio        string `json:"folio" validate:"required"`
	Amount       string `json:"amount" validate:"required"`
	ValidityDate string `json:"validityDate" validate:"required"`
}

// Handle template "envio_cotizacion" from the Whatsapp Business API from Meta
func EnvioCotizacion(w http.ResponseWriter, r *http.Request) {
	var payloadData EnvioCotizacionRequest
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
