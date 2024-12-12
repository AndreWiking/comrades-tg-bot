package gpt

import (
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

const AIkey = "sk-lsQCzJ3tpfnl0Ig0kC1W0k6LL3L2L5P2"

func Test() {

	config := openai.DefaultConfig(AIkey)
	config.BaseURL = "https://api.proxyapi.ru/openai"
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello gtp!",
				},
			},
			MaxCompletionTokens: 1000,
			Temperature:         0.8,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
