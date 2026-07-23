package main

import (
	"error-logging/di"
	"error-logging/router"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Starting Backend Server for error logging")

	container := di.BuildContainer()

	if err := container.Invoke(startServer); err != nil {
		log.Fatal(err)
	}
}

func startServer(r *router.Router) {
	log.Fatal(r.Setup().Run(":8080"))
}
