package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/calmestend/whatsapp-messaging-service/internal/logger"
	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/calmestend/whatsapp-messaging-service/internal/utils"
	"github.com/go-playground/validator/v10"
)

// Struct with parameters for template "encuesta_nps" message
type EncuestaNpsRequest struct {
	models.MessageRequest
	BusinessName string `json:"businessName" validate:"required"`
	Date         string `json:"date" validate:"required"`
	Url          string `json:"url" validate:"required"`
}

// Handle template "encuesta_nps" from the Whatsapp Business API from Meta
func EncuestaNps(w http.ResponseWriter, r *http.Request) {
	var payloadData EncuestaNpsRequest
	err := json.NewDecoder(r.Body).Decode(&payloadData)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON error: %v", err), http.StatusBadRequest)
		logger.Warn("Invalid JSON error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	defer r.Body.Close()

	validate := validator.New()
	err = validate.Struct(payloadData)
	if err != nil {
		errors := err.(validator.ValidationErrors)
		http.Error(w, fmt.Sprintf("Validation error: %v", errors), http.StatusBadRequest)
		logger.Warn("Validation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	url := fmt.Sprintf("https://graph.facebook.com/v22.0/%s/messages", payloadData.From)
	payloadMessage := fmt.Sprintf(`{
	"messaging_product": "whatsapp",
	"to": "%s",
	"type": "template",
	"template": {
		"name": "encuesta_nps_v1",
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
					]
				}
			]
		}
 }`, payloadData.To, payloadData.Url, payloadData.BusinessName, payloadData.Date)

	req, err := http.NewRequest("POST", url, strings.NewReader(payloadMessage))
	if err != nil {
		http.Error(w, fmt.Sprintf("Request creation error: %v", err), http.StatusInternalServerError)
		logger.Warn("Request creation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	req.Header.Set("Authorization", "Bearer "+payloadData.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("HTTP request error: %v", err), http.StatusInternalServerError)
		logger.Warn("HTTP request error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading response error: %v", err), http.StatusInternalServerError)
		logger.Warn("Reading response error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	w.WriteHeader(resp.StatusCode)
	w.Header().Set("Content-type", "application/json")

	parsedBody := utils.ParseJSONBody(respBody)
	logger.LogResponse(resp.StatusCode, r.RequestURI, parsedBody)
	json.NewEncoder(w).Encode(string(respBody))
}

func CreateEncuestaNps(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	err := r.ParseMultipartForm(100 << 20) // 100 MB
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid multipart form error: %v", err), http.StatusBadRequest)
		logger.Warn("Invalid form", "error", err, "uri", r.RequestURI, "method", r.Method)
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

	client := &http.Client{Timeout: 30 * time.Second}

	// Create message template
	templateURL := fmt.Sprintf("https://graph.facebook.com/v23.0/%s/message_templates", payloadData.WbaID)
	templatePayload := `{
		"name": "encuesta_nps_v1",
		"language": "es",
		"category": "utility",
		"components": [
			{
				"type": "header",
				"format": "text",
				"text": "Ayúdanos a mejorar ¿Qué te pareció nuestro servicio?"
			},
			{
				"type": "body",
				"text": "Gracias por visitarnos en {{1}} el {{2}}.\n\nValoramos tus comentarios.\n\nCompleta esta breve encuesta para que sepamos cómo podemos seguir mejorando.",
				"example": {
					"body_text": [["Muebleria Juanito", "1 de enero de 2025"]]
				}
			},
			{
				"type": "BUTTONS",
				"buttons": [
					{
						"type": "url",
						"text": "Completar encuesta",
						"url": "https://sistema.smuebleria.com/{{1}}",
						"example": ["Encuestas/Resolver/Encuestas.aspx?Id=16_249628"]
					}
				]
			}
		]
	}`

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
