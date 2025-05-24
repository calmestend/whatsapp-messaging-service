package main

import (
	"log"
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/routes"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	routes.InitRouter()
	log.Println("Server listening on port 8080")
	http.ListenAndServe(":8080", nil)
}
