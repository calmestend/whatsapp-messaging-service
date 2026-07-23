package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"

	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/logger"
	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/calmestend/whatsapp-messaging-service/internal/utils"
	"github.com/go-playground/validator/v10"
)

// ResponseRecorder is a custom ResponseWriter to capture responses
type ResponseRecorder struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (r *ResponseRecorder) Header() http.Header {
	return r.Headers
}

func (r *ResponseRecorder) Write(data []byte) (int, error) {
	r.Body = append(r.Body, data...)
	return len(data), nil
}

func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
}

func CreateAll(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form instead of regular form
	err := r.ParseMultipartForm(32 << 20) // 32 MB max memory
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
		errors := err.(validator.ValidationErrors)
		http.Error(w, fmt.Sprintf("Validation error: %v", errors), http.StatusBadRequest)
		logger.Warn("Validation error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Rest of your code remains the same...
	// Read local file
	filePath := "template.pdf"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Reading local file %s error: %v", filePath, err), http.StatusInternalServerError)
		logger.Warn("Reading local file error", "file", filePath, "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Create a multipart form with the local file for the requests
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add form fields
	writer.WriteField("phone_id", payloadData.PhoneID)
	writer.WriteField("wba_id", payloadData.WbaID)
	writer.WriteField("token", payloadData.Token)
	writer.WriteField("app_id", payloadData.AppID)

	// Add file part
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath)))
	h.Set("Content-Type", "application/pdf")
	part, err := writer.CreatePart(h)
	if err != nil {
		http.Error(w, fmt.Sprintf("Creating file part error: %v", err), http.StatusInternalServerError)
		logger.Warn("Creating file part error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}
	part.Write(fileData)
	writer.Close()

	// Create a new request with the multipart data
	newReq, err := http.NewRequest("POST", r.URL.String(), bytes.NewReader(requestBody.Bytes()))
	if err != nil {
		http.Error(w, fmt.Sprintf("Creating new request error: %v", err), http.StatusInternalServerError)
		logger.Warn("Creating new request error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	// Copy headers from original request
	for key, values := range r.Header {
		for _, value := range values {
			newReq.Header.Add(key, value)
		}
	}

	// Set the multipart content type
	newReq.Header.Set("Content-Type", writer.FormDataContentType())

	// List of template creation functions to call
	createFunctions := []struct {
		name string
		fn   func(http.ResponseWriter, *http.Request)
	}{
		{"pagos_atrasados", CreatePagosAtrasados},
		{"encuesta_nps", CreateEncuestaNps},
		{"pedido_confirmado", CreatePedidoConfirmado},
		{"envio_cotizacion", CreateEnvioCotizacion},
		{"envio_compra", CreateEnvioCompra},
		{"cierre_diario", CreateCierreDiario},
		{"pago_por_vencer", CreatePagoPorVencer},
		{"saldo_vencido", CreateSaldoVencido},
		{"envio_factura", CreateEnvioFactura},
	}

	results := make(map[string]any)
	errors := make(map[string]string)

	// Call each create function
	for _, createFunc := range createFunctions {
		// Create a custom ResponseWriter to capture the response
		recorder := &ResponseRecorder{
			StatusCode: http.StatusOK,
			Headers:    make(http.Header),
		}

		// Create a fresh request for each function call
		freshReq, err := http.NewRequest("POST", r.URL.String(), bytes.NewReader(requestBody.Bytes()))
		if err != nil {
			errors[createFunc.name] = fmt.Sprintf("Error creating fresh request: %v", err)
			continue
		}

		// Copy headers
		for key, values := range newReq.Header {
			for _, value := range values {
				freshReq.Header.Add(key, value)
			}
		}

		// Call the create function
		createFunc.fn(recorder, freshReq)

		// Store the result
		if recorder.StatusCode >= 200 && recorder.StatusCode < 300 {
			results[createFunc.name] = map[string]any{
				"status": "success",
				"code":   recorder.StatusCode,
				"body":   string(recorder.Body),
			}
		} else {
			errors[createFunc.name] = string(recorder.Body)
		}
	}

	var message string
	var statusCode int

	switch {
	case len(results) > 0 && len(errors) > 0:
		message = "Some templates failed to create"
		statusCode = http.StatusPartialContent // 206
	case len(errors) > 0 && len(results) == 0:
		message = "All templates failed to create"
		statusCode = http.StatusInternalServerError // 500
	default:
		message = "All templates created successfully"
		statusCode = http.StatusOK // 200
	}

	// Build the summary response as a real map and marshal it,
	// instead of hand-building a JSON string with %v (which produced
	// invalid JSON for the nested "results"/"errors" maps).
	response := map[string]any{
		"message":    message,
		"successful": len(results),
		"failed":     len(errors),
		"results":    results,
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Marshaling response error: %v", err), http.StatusInternalServerError)
		logger.Warn("Marshaling response error", "error", err, "uri", r.RequestURI, "method", r.Method)
		return
	}

	parsedBody := utils.ParseJSONBody(responseBody)
	logger.LogResponse(statusCode, r.RequestURI, parsedBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}
