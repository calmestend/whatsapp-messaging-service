package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"github.com/calmestend/whatsapp-messaging-service/internal/logger"
	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/calmestend/whatsapp-messaging-service/internal/utils"
	"github.com/go-playground/validator/v10"
)

// Struct with parameters for template "cierre_diario" message
type CierreDiarioRequest struct {
	models.MessageRequest
	CustomerName string `validate:"required"`
	BusinessName string `json:"businessName" validate:"required"`
	Date         string `validate:"required"`
	Hour         string `validate:"required"`
}

// Handle template "cierre_diario" from the Whatsapp Business API from Meta
func CierreDiario(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid multipart form error: %v", err), http.StatusBadRequest)
		logger.Warn("Invalid multipart form error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	payloadData := CierreDiarioRequest{
		MessageRequest: models.MessageRequest{
			To:    r.FormValue("to"),
			From:  r.FormValue("from"),
			Token: r.FormValue("token"),
		},
		CustomerName: r.FormValue("customerName"),
		BusinessName: r.FormValue("businessName"),
		Date:         r.FormValue("date"),
		Hour:         r.FormValue("hour"),
	}
	validate := validator.New()
	err = validate.Struct(payloadData)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		http.Error(w, fmt.Sprintf("Validation error: %v", errors), http.StatusBadRequest)
		logger.Warn("Validation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Read file into buffer
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Missing or invalid file error: %v", err), http.StatusBadRequest)
		logger.Warn("Missing or invalid file error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	defer file.Close()

	var fileBuffer bytes.Buffer
	_, err = io.Copy(&fileBuffer, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading file error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading file error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Upload PDF to Facebook media endpoint
	uploadURL := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/media", payloadData.From)
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.WriteField("messaging_product", "whatsapp")

	// Create part with PDF MIME
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="cierre_diario.pdf"`)
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(h)
	if err != nil {
		http.Error(w, fmt.Sprintf("Create part with PDF MIME error: %v", err), http.StatusInternalServerError)
		logger.Warn("Create part with PDF MIME error", "error", err, "uri", r.RequestURI, "method", r.Method)
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
		http.Error(w, fmt.Sprintf("Upload error: %v", err), http.StatusInternalServerError)
		logger.Warn("Upload error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	defer uploadResp.Body.Close()

	// Read upload response
	uploadBody, _ := io.ReadAll(uploadResp.Body)

	var uploadResult struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(uploadBody, &uploadResult); err != nil {
		http.Error(w, fmt.Sprintf("Parsing media ID error: %v", err), http.StatusInternalServerError)
		logger.Warn("Parsing media ID error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Send template payload to whatsapp endpoint
	msgURL := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", payloadData.From)
	payloadMessage := fmt.Sprintf(`{
	"messaging_product": "whatsapp",
	"to": "%s",
	"type": "template",
	"template": {
		"name": "cierre_diario_v1",
		"language": { "code": "es" },
		"components": [
			{
				"type": "header",
				"parameters": [
					{
						"type": "document",
						"document": {
							"id": "%s",
							"filename": "cierre_diario.pdf"
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
					]
				}
			]
		}
 }`, payloadData.To, uploadResult.ID, payloadData.BusinessName, payloadData.Date, payloadData.Hour, payloadData.CustomerName)

	msgReq, err := http.NewRequest("POST", msgURL, strings.NewReader(payloadMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Request creation error: %v", err), http.StatusInternalServerError)
		logger.Warn("Request creation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	msgReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	msgReq.Header.Set("Content-Type", "application/json")

	client = &http.Client{Timeout: 10 * time.Second}
	msgResp, err := client.Do(msgReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP request error: %v", err), http.StatusInternalServerError)
		logger.Warn("HTTP request error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	defer msgResp.Body.Close()

	body, err := io.ReadAll(msgResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading response error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading response error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	w.WriteHeader(msgResp.StatusCode)
	w.Header().Set("Content-Type", "application/json")

	parsedBody := utils.ParseJSONBody(body)
	logger.LogResponse(msgResp.StatusCode, r.RequestURI, parsedBody)
	json.NewEncoder(w).Encode(parsedBody)
}

func CreateCierreDiario(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid multipart form error: %v", err), http.StatusBadRequest)
		logger.Warn("Invalid multipart form error", "error", err, "uri", r.RequestURI, "method", r.Method)
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
		errors := err.(validator.ValidationErrors)
		http.Error(w, fmt.Sprintf("Validation error: %v", errors), http.StatusBadRequest)
		logger.Warn("Validation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("Missing or invalid file error: %v", err), http.StatusInternalServerError)
		logger.Warn("Missing or invalid file error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	defer file.Close()

	var fileBuffer bytes.Buffer

	_, err = io.Copy(&fileBuffer, file)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading file error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading file error", "error", err, "uri", r.RequestURI, "method", r.Method)
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
		logger.Warn("Upload session request creation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	uploadSessionReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	uploadSessionReq.Header.Set("Content-Type", "application/json")

	uploadSessionResp, err := client.Do(uploadSessionReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload session request error: %v", err), http.StatusInternalServerError)
		logger.Warn("Upload session request error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	defer uploadSessionResp.Body.Close()

	uploadSessionBody, err := io.ReadAll(uploadSessionResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading upload session response error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading upload session response error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	uploadSessionBodyParsed := utils.ParseJSONBody(uploadSessionBody)

	if uploadSessionResp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Upload session failed with status: %s, body: %s", uploadSessionResp.Status, string(uploadSessionBody)), uploadSessionResp.StatusCode)
		logger.LogResponse(uploadSessionResp.StatusCode, r.RequestURI, uploadSessionBodyParsed)
		return
	}

	var uploadSessionResult struct {
		ID string `json:"id"`
	}

	err = json.Unmarshal(uploadSessionBody, &uploadSessionResult)
	if err != nil {
		http.Error(w, fmt.Sprintf("Parsing upload session ID error: %v", err), http.StatusInternalServerError)
		logger.Warn("Parsing upload session ID error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Upload the file
	uploadFileURL := fmt.Sprintf("https://graph.facebook.com/v23.0/%s", uploadSessionResult.ID)

	var uploadBuffer bytes.Buffer
	writer := multipart.NewWriter(&uploadBuffer)

	// Create part with PDF MIME
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileHeader.Filename))
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(h)
	if err != nil {
		http.Error(w, fmt.Sprintf("Create part with PDF MIME error: %v", err), http.StatusInternalServerError)
		logger.Warn("Create part with PDF MIME error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	io.Copy(part, &fileBuffer)
	writer.Close()

	uploadFileReq, err := http.NewRequest("POST", uploadFileURL, &uploadBuffer)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusInternalServerError)
		logger.Warn("Upload file error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	uploadFileReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	uploadFileReq.Header.Set("Content-Type", writer.FormDataContentType())

	uploadFileResp, err := client.Do(uploadFileReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload file error: %v", err), http.StatusInternalServerError)
		logger.Warn("Upload file error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	defer uploadFileResp.Body.Close()

	uploadFileBody, err := io.ReadAll(uploadFileResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading upload file response error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading upload file response error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	uploadFileBodyParsed := utils.ParseJSONBody(uploadFileBody)

	if uploadFileResp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("File upload failed with status: %s, body: %s", uploadFileResp.Status, string(uploadFileBody)), uploadFileResp.StatusCode)
		logger.LogResponse(uploadSessionResp.StatusCode, r.RequestURI, uploadFileBodyParsed)
		return
	}

	var uploadFileResult struct {
		H string `json:"h"`
	}
	if err := json.Unmarshal(uploadFileBody, &uploadFileResult); err != nil {
		http.Error(w, fmt.Sprintf("Parsing upload file handle failed: %v", err), http.StatusInternalServerError)
		logger.Warn("Parsing upload file handle error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Create message template
	templateURL := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/message_templates", payloadData.WbaID)
	templatePayload := fmt.Sprintf(`{
		"name": "cierre_diario_v1",
		"language": "es",
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
				"text": "Hola {{4}},\nTe compartimos el *cierre diario* de la empresa *{{1}}* que se genero el dia *{{2}} *a la hora *{{3}}*\nSi deseas tener mas detalles consulta el reporte de bancos o cuentas en SMuebleria.",
				"example": {
					"body_text": [["Muebleria X", "06/06/2025", "10:35 pm", "Alejandro Velasco"]]
				}
			}
		]
	}`, uploadFileResult.H)

	templateReq, err := http.NewRequest("POST", templateURL, strings.NewReader(templatePayload))
	if err != nil {
		http.Error(w, fmt.Sprintf("Template request creation error: %v", err), http.StatusInternalServerError)
		logger.Warn("Template request creation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	templateReq.Header.Set("Authorization", "Bearer "+payloadData.Token)
	templateReq.Header.Set("Content-Type", "application/json")

	templateResp, err := client.Do(templateReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template request error: %v", err), http.StatusInternalServerError)
		logger.Warn("Template request error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	defer templateResp.Body.Close()

	templateBody, err := io.ReadAll(templateResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading template response error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading template response error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	w.WriteHeader(templateResp.StatusCode)
	w.Header().Set("Content-Type", "application/json")

	parsedBody := utils.ParseJSONBody(templateBody)
	logger.LogResponse(templateResp.StatusCode, r.RequestURI, parsedBody)
	w.Write(templateBody)
}
