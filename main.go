package main

import (
	"fmt"
	"log"

	"github.com/Lanrey-waju/gator.git/internal/config"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	err = cfg.SetUser("Abdulmumin")
	if err != nil {
		log.Fatalf("Error setting user: %v", err)
	}
	cfg, err = config.Read()
	if err != nil {
		log.Fatalf("Error reading updated config: %v", err)
	}
	fmt.Printf("DB URL: %s\nCurrent User: %s\n", cfg.DBUrl, cfg.CurrentUserName)
}
