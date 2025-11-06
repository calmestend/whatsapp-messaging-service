package main

import (
	"fmt"
	"net/http"

	"github.com/calmestend/whatsapp-messaging-service/internal/logger"
	"github.com/calmestend/whatsapp-messaging-service/internal/routes"
)

func main() {
	file := logger.Init("service.log")
	defer file.Close()

	const port int32 = 8080

	logger.Info("Starting server")
	routes.InitRouter()

	logger.Info("Server started", "port", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
