// Package routes implements the router initialization and behaviour.
package routes

import (
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/handlers"
)

// @TODO: Create tests
// @TODO: Add "links" handling
// @TODO: Add "documents" handling
// @TODO: Add "surveys" handling
// @TODO: Add "media" handling

// Init router
func InitRouter() {
	http.HandleFunc("POST /api/v1/template/pagos_atrasados", handlers.PagosAtrasados)
	http.HandleFunc("POST /api/v1/template/encuesta_nps", handlers.EncuestaNps)
	http.HandleFunc("POST /api/v1/template/pedido_confirmado", handlers.PedidoConfirmado)
	http.HandleFunc("POST /api/v1/template/envio_cotizacion", handlers.EnvioCotizacion)
}
