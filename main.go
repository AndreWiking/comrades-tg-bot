package main

import (
	"ComradesTG/bot"
	"ComradesTG/parser/tg"
	"ComradesTG/settings"
	"log"
	"os"
)

func workBot() {

	myBot := bot.NewBot()
	myBot.RunUpdates()

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
		log.Fatal("Unknown prog type:", argsWithoutProg[0])
	}
}

/*

ssh root@46.17.41.227

sudo systemctl restart nginx

7541929739:AAFylnUcAeDvSueJGIGQ5kAfow4nEw7P-Oc

ssh root@46.17.41.227
scp -r /Users/andrewiking/GolandProjects/ComradesTG root@46.17.41.227:/root/
go build .
systemctl restart ComradesTG
systemctl status ComradesTG
systemctl stop ComradesTG

psql -h <REMOTE HOST> -p <REMOTE PORT> -U <DB_USER> <DB_NAME>

psql -h 46.17.41.227 -U super_admin postgres

su - postgres
psql

systemctl start ComradesTG
systemctl status ComradesTG


systemctl status postgres


https://t.me/find_comrade_bot?start=ya1

*/
