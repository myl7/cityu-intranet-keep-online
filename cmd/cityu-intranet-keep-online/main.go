package main

import (
	"os"

	"github.com/myl7/cityu-intranet-keep-online"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Logging
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logLevel, err := log.ParseLevel(level)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)

	auth := cityu.NewAuthFromEnv()
	err = auth.Login()
	if err != nil {
		log.Fatal(err)
	}
}
