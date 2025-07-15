package handlers

import (
	"fmt"
	"github.com/calmestend/whatsapp-messaging-service/internal/models"
	"github.com/go-playground/validator/v10"
	"net/http"
)

func CreateAll(w http.ResponseWriter, r *http.Request) {
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

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing or invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

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

		// Reset the file position for each function call
		file.Seek(0, 0)

		// Call the create function
		createFunc.fn(recorder, r)

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
