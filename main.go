package main

import (
	"log"
	"powermate-integration-testing/configuration"
)

func main() {
	config := configuration.Load()
	log.Println(config.Host)
}
