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
	"strings"
	"time"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
)

// Struct with parameters for template "envio_compra" message
type EnvioCompraRequest struct {
	models.MessageRequest
	SupplierName string `json:"supplierName" validate:"required"`
	BusinessName string `json:"businessName" validate:"required"`
}

// Handle template "envio_compra" from the Whatsapp Business API from Meta
func EnvioCompra(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	payloadData := EnvioCompraRequest{
		MessageRequest: models.MessageRequest{
			To:    r.FormValue("to"),
			From:  r.FormValue("from"),
			Token: r.FormValue("token"),
		},
		BusinessName: r.FormValue("businessName"),
		SupplierName: r.FormValue("supplierName"),
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
	h.Set("Content-Disposition", `form-data; name="file"; filename="envio_compra.pdf"`)
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
	uploadReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
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
		"name": "envio_compra",
		"language": { "code": "es_MX" },
		"components": [
			{
				"type": "header",
				"parameters": [
					{
						"type": "document",
						"document": {
							"id": "%s",
							"filename": "envio_compra.pdf"
						}
					}
				]
			},
				{
					"type": "body",
					"parameters": [
						{ "type": "text", "text": "%s"},
						{ "type": "text", "text": "%s"},
					]
				}
			]
		}
 }`, payloadData.To, uploadResult.ID, payloadData.SupplierName, payloadData.BusinessName)

	msgReq, err := http.NewRequest("POST", msgURL, strings.NewReader(payloadMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Request creation error: %v", err), http.StatusInternalServerError)
		return
	}

	msgReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
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

func CreateEnvioCompra(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, "Invalid multipart form", http.StatusBadRequest)
		return
	}

	payloadData := models.TemplateRequest{
		PhoneID: r.FormValue("phone_id"),
		WbaID:   r.FormValue("wba_id"),
		Token:   r.FormValue("token"),
		AppID:   r.FormValue("app_id"),
	}

	validate := validator.New()
	err = validate.Struct(payloadData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %s", err), http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
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

	client := &http.Client{Timeout: 30 * time.Second}

	// Create upload session
	uploadSessionURL := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/uploads", payloadData.AppID)
	uploadSessionPayload := fmt.Sprintf(`{
		"file_length": %d,
		"file_name": "%s",
		"file_type": "application/pdf",
		"session_type": "attachment"
	}`, fileBuffer.Len(), fileHeader.Filename)

	uploadSessionReq, err := http.NewRequest("POST", uploadSessionURL, strings.NewReader(uploadSessionPayload))
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload session request creation error: %v", err), http.StatusInternalServerError)
		return
	}

	uploadSessionReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	uploadSessionReq.Header.Set("Content-Type", "application/json")

	uploadSessionResp, err := client.Do(uploadSessionReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload session request error: %v", err), http.StatusInternalServerError)
		return
	}
	defer uploadSessionResp.Body.Close()

	uploadSessionBody, err := io.ReadAll(uploadSessionResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading upload session response error: %v", err), http.StatusInternalServerError)
		return
	}

	if uploadSessionResp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Upload session failed with status: %s, body: %s", uploadSessionResp.Status, string(uploadSessionBody)), uploadSessionResp.StatusCode)
		return
	}

	var uploadSessionResult struct {
		ID string `json:"id"`
	}

	err = json.Unmarshal(uploadSessionBody, &uploadSessionResult)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parsing upload session ID failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Upload the file
	uploadFileURL := fmt.Sprintf("https://graph.facebook.com/v23.0/%s", uploadSessionResult.ID)

	var uploadBuffer bytes.Buffer
	writer := multipart.NewWriter(&uploadBuffer)

	// Create file part
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileHeader.Filename))
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(h)
	if err != nil {
		http.Error(w, fmt.Sprintf("CreatePart error: %v", err), http.StatusInternalServerError)
		return
	}
	io.Copy(part, &fileBuffer)
	writer.Close()

	uploadFileReq, err := http.NewRequest("POST", uploadFileURL, &uploadBuffer)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload file request creation error: %v", err), http.StatusInternalServerError)
		return
	}

	uploadFileReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	uploadFileReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadFileResp, err := client.Do(uploadFileReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload file request error: %v", err), http.StatusInternalServerError)
		return
	}
	defer uploadFileResp.Body.Close()

	uploadFileBody, err := io.ReadAll(uploadFileResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading upload file response error: %v", err), http.StatusInternalServerError)
		return
	}

	if uploadFileResp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("File upload failed with status: %s, body: %s", uploadFileResp.Status, string(uploadFileBody)), uploadFileResp.StatusCode)
		return
	}

	var uploadFileResult struct {
		H string `json:"h"`
	}
	if err := json.Unmarshal(uploadFileBody, &uploadFileResult); err != nil {
		http.Error(w, fmt.Sprintf("Parsing upload file handle failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Create message template
	templateURL := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/message_templates", payloadData.WbaID)
	templatePayload := fmt.Sprintf(`{
		"name": "envio_compra_test",
		"language": "es_MX",
		"category": "utility",
		"components": [
			{
				"type": "header",
				"format": "document",
				"example": {
					"header_handle": [
						"%s"
					]
				}
			},
			{
				"type": "body",
				"text": "Estimado Proveedor {{1}}\nLa empresa *{{2}}* ha generado una nueva orden de compra, todos los detalles los encontrará en el PDF adjunto.\nLe agradecemos si nos confirma de recibida y apoya con el seguimiento.",
				"example": {
					"body_text": [["Muebles X", "Muebleria Y"]]
				}
			}
		]
	}`, uploadFileResult.H)

	templateReq, err := http.NewRequest("POST", templateURL, strings.NewReader(templatePayload))
	if err != nil {
		http.Error(w, fmt.Sprintf("Template request creation error: %v", err), http.StatusInternalServerError)
		return
	}

	templateReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	templateReq.Header.Set("Content-Type", "application/json")

	templateResp, err := client.Do(templateReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template request error: %v", err), http.StatusInternalServerError)
		return
	}
	defer templateResp.Body.Close()

	templateBody, err := io.ReadAll(templateResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading template response error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(templateResp.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(templateBody)
}
