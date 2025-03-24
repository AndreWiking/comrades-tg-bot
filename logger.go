package main

import (
	"ComradesTG/settings"
	"log"
	"os"
)

func SetLogger() *os.File {

	logFile, err := os.OpenFile(settings.LogFilePath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	log.SetOutput(logFile)

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Println("Session started")

	return logFile
}
