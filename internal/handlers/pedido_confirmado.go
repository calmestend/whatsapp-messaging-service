package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
)

// @TODO: Create utils function to handle payloadData validations

// Struct with parameters for template "pedido_confirmado" message
type PedidoConfirmadoRequest struct {
	models.MessageRequest
	CustomerName     string `json:"customerName" validate:"required"`
	OrderType        string `json:"orderType" validate:"required"`
	Folio            string `json:"folio" validate:"required"`
	NumberOfProducts int    `json:"numberOfProducts" validate:"required"`
	Amount           string `json:"amount" validate:"required"`
	SellerName       string `json:"sellerName" validate:"required"`
	Url              string `json:"url" validate:"required"`
}

// Handle template "pedido_confirmado" from the Whatsapp Business API from Meta
func PedidoConfirmado(w http.ResponseWriter, r *http.Request) {
	var token string = os.Getenv("WBA_TOKEN")
	if token == "" {
		fmt.Println(token)
		http.Error(w, "Missing WBA_TOKEN in environment", http.StatusInternalServerError)
		return
	}

	var payloadData PedidoConfirmadoRequest
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

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", payloadData.From)
	payloadMessage := fmt.Sprintf(`{
	"messaging_product": "whatsapp",
	"to": "%s",
	"type": "template",
	"template": {
		"name": "pedido_confirmado",
		"language": { "code": "es" },
		"components": [
				{
					"type": "button",
					"sub_type": "url",
					"index": 0,
					"parameters": [
						{
							"type": "text",
							"text": "%s"
						}
					]
				},
				{
					"type": "body",
					"parameters": [
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%d"},
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"}
					]
				}
			]
		}
 }`, payloadData.To, payloadData.Url, payloadData.CustomerName, payloadData.OrderType, payloadData.Folio, payloadData.NumberOfProducts, payloadData.Amount, payloadData.SellerName)

	req, err := http.NewRequest("POST", url, strings.NewReader(payloadMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Request creation error: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP request error: %v", err), http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading response error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Header().Set("Content-type", "application/json")
	fmt.Fprintln(w, "Message sent")
	json.NewEncoder(w).Encode(string(respBody))
}
