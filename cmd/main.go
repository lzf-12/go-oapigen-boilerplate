package main

import (
	"log"
	"oapi-to-rest/api"
)

func main() {

	server := api.NewServer()
	server.RegisterRoutes()
	server.PrintRoutes()
	log.Fatal(server.Start(":8080"))
}
