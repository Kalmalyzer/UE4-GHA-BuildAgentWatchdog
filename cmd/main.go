package main

import (
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	watchdog "github.com/falldamagestudio/UE4-GHA-BuildAgentWatchdog"
)

func main() {
	funcframework.RegisterHTTPFunction("/", watchdog.RunWatchdog)

	// Use PORT environment variable, or default to 8080.
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
