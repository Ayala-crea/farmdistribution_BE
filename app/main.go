package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	// "gobizdevelop/config"
	"app_farm_be/routes"
)

func main() {
	router := routes.InitializeRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
