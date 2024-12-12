package main

import (
	"log"
	"os"
)

const logFileName = "logs.txt"

func SetLogger() {

	logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()
	defer log.Println("Session finished")

	log.SetOutput(logFile)

	log.SetFlags(log.Lshortfile | log.LstdFlags)

	log.Println("Session started")
}
