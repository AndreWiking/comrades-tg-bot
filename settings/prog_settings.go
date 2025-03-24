package settings

type ProgType int

const (
	TgBotProgType ProgType = iota
	TgParserProgType
)

var ProgTypeName = map[ProgType]string{
	TgBotProgType:    "tg-bot",
	TgParserProgType: "tg-parser",
}

const LogFilePath = "logs.txt"
