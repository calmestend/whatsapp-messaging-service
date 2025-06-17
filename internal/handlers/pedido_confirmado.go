package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"strings"
	"time"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
)

// Struct with parameters for template "pedido_confirmado" message
type PedidoConfirmadoRequest struct {
	models.MessageRequest
	CustomerName     string `json:"customerName" validate:"required"`
	BusinessName     string `json:"businessName" validate:"required"`
	Folio            string `json:"folio" validate:"required"`
	NumberOfProducts string `json:"numberOfProducts" validate:"required"`
	Amount           string `json:"amount" validate:"required"`
	SellerName       string `json:"sellerName" validate:"required"`
}

// Handle template "pedido_confirmado" from the Whatsapp Business API from Meta
func PedidoConfirmado(w http.ResponseWriter, r *http.Request) {
	var token string = os.Getenv("WBA_TOKEN")
	if token == "" {
		http.Error(w, "Missing WBA_TOKEN in environment", http.StatusInternalServerError)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	payloadData := PedidoConfirmadoRequest{
		MessageRequest: models.MessageRequest{
			To:   r.FormValue("to"),
			From: r.FormValue("from"),
		},
		CustomerName:     r.FormValue("customerName"),
		BusinessName:     r.FormValue("businessName"),
		Folio:            r.FormValue("folio"),
		Amount:           r.FormValue("amount"),
		NumberOfProducts: r.FormValue("numberOfProducts"),
		SellerName:       r.FormValue("sellerName"),
	}
	validate := validator.New()
	err = validate.Struct(payloadData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %s", err), http.StatusBadRequest)
		return
	}

	// Read file into buffer
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing or invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	var fileBuffer bytes.Buffer
	_, err = io.Copy(&fileBuffer, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading file: %v", err), http.StatusInternalServerError)
		return
	}

	// Upload PDF to Facebook media endpoint
	uploadURL := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/media", payloadData.From)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("messaging_product", "whatsapp")

	// Create part with PDF MIME
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="pedido_confirmado.pdf"`)
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(h)
	if err != nil {
		http.Error(w, fmt.Sprintf("CreatePart error: %v", err), http.StatusInternalServerError)
		return
	}
	io.Copy(part, &fileBuffer)
	writer.Close()

	client := &http.Client{Timeout: 10 * time.Second}
	uploadReq, _ := http.NewRequest("POST", uploadURL, &b)
	uploadReq.Header.Set("Authorization", "Bearer "+token)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadResp, err := client.Do(uploadReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer uploadResp.Body.Close()

	// Read upload response
	uploadBody, _ := io.ReadAll(uploadResp.Body)
	log.Printf("Upload response status: %s", uploadResp.Status)
	log.Printf("Upload response body: %s", string(uploadBody))

	var uploadResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResult); err != nil {
		http.Error(w, fmt.Sprintf("Parsing media ID failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send template payload to whatsapp endpoint
	msgURL := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", payloadData.From)
	payloadMessage := fmt.Sprintf(`{
	"messaging_product": "whatsapp",
	"to": "%s",
	"type": "template",
	"template": {
		"name": "pedido_confirmado",
		"language": { "code": "es" },
		"components": [
			{
				"type": "header",
				"parameters": [
					{
						"type": "document",
						"document": {
							"id": "%s",
							"filename": "pedido_confirmado.pdf"
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
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"}
					]
				}
			]
		}
 }`, payloadData.To, uploadResult.ID, payloadData.CustomerName, payloadData.BusinessName, payloadData.Folio, payloadData.NumberOfProducts, payloadData.Amount, payloadData.SellerName)

	msgReq, err := http.NewRequest("POST", msgURL, strings.NewReader(payloadMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Request creation error: %v", err), http.StatusInternalServerError)
		return
	}

	msgReq.Header.Set("Authorization", "Bearer "+token)
	msgReq.Header.Set("Content-Type", "application/json")

	client = &http.Client{Timeout: 10 * time.Second}
	msgResp, err := client.Do(msgReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP request error: %v", err), http.StatusInternalServerError)
		return
	}

	defer msgResp.Body.Close()

	body, err := io.ReadAll(msgResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading response error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(msgResp.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(string(body))
}
