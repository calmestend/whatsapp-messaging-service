// Package routes implements the router initialization and behaviour.
package routes

import (
	"log"
	"net/http"
)

// Init router
func InitRouter() {
	http.HandleFunc("POST /api/v1/template/pagos_atrasados", func(http.ResponseWriter, *http.Request) {
		log.Println("pagos_atrasados")
	})
	http.HandleFunc("POST /api/v1/template/encuesta_nps", func(http.ResponseWriter, *http.Request) {
		log.Println("encuesta_nps")
	})
	http.HandleFunc("POST /api/v1/template/pedido_confirmado", func(http.ResponseWriter, *http.Request) {
		log.Println("pedido_confirmado")
	})
	http.HandleFunc("POST /api/v1/template/envio_cotizacion", func(http.ResponseWriter, *http.Request) {
		log.Println("envio_cotizacion")
	})
}
