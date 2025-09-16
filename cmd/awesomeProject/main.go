package main

import (
	"log"

	"lab1-design-backend/internal/api"
)

func main() {
	log.Println("Application start!")
	api.StartServer()
	log.Println("Application terminated!")
}
