package main

import (
	"ComradesTG/bot"
	"ComradesTG/parser/tg"
	"ComradesTG/settings"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func workBot() {
	myBot := bot.NewBot()
	myBot.RunUpdates()
}

func init() {
	if err := godotenv.Load(settings.EnvFilePath); err != nil {
		log.Fatalln("No .env file found")
	}
}

func main() {
	logFile := SetLogger()
	defer logFile.Close()
	defer log.Println("Session finished")

	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		log.Fatal("Set prog type")
	}
	switch argsWithoutProg[0] {
	case settings.ProgTypeName[settings.TgBotProgType]:
		workBot()
	case settings.ProgTypeName[settings.TgParserProgType]:
		tg.Parse()
	default:
		log.Fatalln("Unknown prog type:", argsWithoutProg[0])
	}
}
