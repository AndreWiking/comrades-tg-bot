package main

import (
	"log"
	"os"
)

const logFileName = "logs.txt"

func SetLogger() *os.File {

	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	log.SetOutput(logFile)

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Println("Session started")

	return logFile
}
