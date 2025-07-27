package handlers

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
	"net/http"
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
		http.Error(w, fmt.Sprintf("Validation error: %s", err), http.StatusBadRequest)
		return
	}

	// Rest of your code remains the same...
	// Read local file
	filePath := "template.pdf"
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading local file %s: %v", filePath, err), http.StatusInternalServerError)
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
		http.Error(w, fmt.Sprintf("Error creating file part: %v", err), http.StatusInternalServerError)
		return
	}
	part.Write(fileData)
	writer.Close()

	// Create a new request with the multipart data
	newReq, err := http.NewRequest("POST", r.URL.String(), bytes.NewReader(requestBody.Bytes()))
	if err != nil {
		http.Error(w, fmt.Sprintf("Error creating new request: %v", err), http.StatusInternalServerError)
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

	// Prepare the response
	response := map[string]any{
		"results": results,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		w.WriteHeader(http.StatusPartialContent) // 206 for partial success
	} else {
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")

	// Send a summary response
	responseBody := fmt.Sprintf(`{
		"message": "Template creation completed",
		"successful": %d,
		"failed": %d,
		"results": %v
	}`, len(results), len(errors), response)

	w.Write([]byte(responseBody))
}
