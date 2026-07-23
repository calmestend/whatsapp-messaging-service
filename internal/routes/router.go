// Package routes implements the router initialization and behaviour.
package routes

import (
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/handlers"
)

// Init router
func InitRouter() {
	// Create Templates
	http.HandleFunc("POST /api/v1/template/create/pagos_atrasados", handlers.CreatePagosAtrasados)
	http.HandleFunc("POST /api/v1/template/create/encuesta_nps", handlers.CreateEncuestaNps)
	http.HandleFunc("POST /api/v1/template/create/pedido_confirmado", handlers.CreatePedidoConfirmado)
	http.HandleFunc("POST /api/v1/template/create/envio_cotizacion", handlers.CreateEnvioCotizacion)
	http.HandleFunc("POST /api/v1/template/create/envio_compra", handlers.CreateEnvioCompra)
	http.HandleFunc("POST /api/v1/template/create/cierre_diario", handlers.CreateCierreDiario)
	http.HandleFunc("POST /api/v1/template/create/pago_por_vencer", handlers.CreatePagoPorVencer)
	http.HandleFunc("POST /api/v1/template/create/saldo_vencido", handlers.CreateSaldoVencido)
	http.HandleFunc("POST /api/v1/template/create/envio_factura", handlers.CreateEnvioFactura)

	// Create All Templates
	http.HandleFunc("POST /api/v1/templates/create", handlers.CreateAll)

	// Templates
	http.HandleFunc("POST /api/v1/template/pagos_atrasados", handlers.PagosAtrasados)
	http.HandleFunc("POST /api/v1/template/encuesta_nps", handlers.EncuestaNps)
	http.HandleFunc("POST /api/v1/template/pedido_confirmado", handlers.PedidoConfirmado)
	http.HandleFunc("POST /api/v1/template/envio_cotizacion", handlers.EnvioCotizacion)
	http.HandleFunc("POST /api/v1/template/envio_compra", handlers.EnvioCompra)
	http.HandleFunc("POST /api/v1/template/cierre_diario", handlers.CierreDiario)
	http.HandleFunc("POST /api/v1/template/pago_por_vencer", handlers.PagoPorVencer)
	http.HandleFunc("POST /api/v1/template/saldo_vencido", handlers.SaldoVencido)
	http.HandleFunc("POST /api/v1/template/envio_factura", handlers.EnvioFactura)
}
