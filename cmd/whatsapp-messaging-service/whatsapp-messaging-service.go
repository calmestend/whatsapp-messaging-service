package main

import (
	"log"
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/routes"
)

func main() {
	routes.InitRouter()
	log.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
