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

const (
	TgApiKeyName  = "TG_API_KEY"
	GptApiKeyName = "GPT_API_KEY"

	EnvFilePath = "config.env"
	LogFilePath = "logs.txt"
)
