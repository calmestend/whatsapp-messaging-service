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
// @TODO: Change EnvioCotizacionRequest.Amount to int and hardcode "$"

// @NOTE: I need to manage the amounts of bytes generated for the pdf without saving it and pass it across the network with a multipart/form-data.
// > I need to make the pdf and sort of bytes ([]bytes) and send it in the body json from the pdf generator service,
// > then i process it from this microservice (whatsapp messaging service) and store it using the Meta's Cloud API
// > If the pdf is saved successfully i need to take back the id and now use it as a parameter to my message template

// @IMPORTANT: The "link" parameter is temporal, it will be removed when the feature is finished

// Struct with parameters for template "envio_cotizacion" message
type EnvioCotizacionRequest struct {
	models.MessageRequest
	CustomerName string `json:"customerName" validate:"required"`
	BusinessName string `json:"businessName" validate:"required"`
	Folio        string `json:"folio" validate:"required"`
	Amount       string `json:"amount" validate:"required"`
	Date         string `json:"date" validate:"required"`
}

// Handle template "envio_cotizacion" from the Whatsapp Business API from Meta
func EnvioCotizacion(w http.ResponseWriter, r *http.Request) {
	var token string = os.Getenv("WBA_TOKEN")
	if token == "" {
		fmt.Println(token)
		http.Error(w, "Missing WBA_TOKEN in environment", http.StatusInternalServerError)
		return
	}

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

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", payloadData.From)
	payloadMessage := fmt.Sprintf(`{
	"messaging_product": "whatsapp",
	"to": "%s",
	"type": "template",
	"template": {
		"name": "envio_cotizacion",
		"language": { "code": "es_MX" },
		"components": [
				{
					"type": "header",
					"parameters": [
						{
							"type": "document",
							"document": {
								"link": "https://pdfobject.com/pdf/sample.pdf",
								"filename": "cotizacion.pdf"
							}
						}
					]						
				},
				{
					"type": "body",
					"parameters": [
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"}
					]
				}
			]
		}
 }`, payloadData.To, payloadData.BusinessName, payloadData.Folio, payloadData.Amount, payloadData.Date, payloadData.CustomerName)

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
