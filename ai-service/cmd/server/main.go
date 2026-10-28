package main

import (
	"ai-service/internal/app"
	"log"
	"os"
)

func main() {
	if err := app.Run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
