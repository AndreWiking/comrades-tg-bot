package gpt

import (
	"context"
	"encoding/json"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
)

const (
	AIkey       = "sk-lsQCzJ3tpfnl0Ig0kC1W0k6LL3L2L5P2"
	BaseUrl     = "https://api.proxyapi.ru/openai/v1"
	maxTokens   = 1000
	temperature = 0.8
)

type Client struct {
	client *openai.Client
}

var Connection Client

func NewClient() {
	config := openai.DefaultConfig(AIkey)
	config.BaseURL = BaseUrl
	Connection.client = openai.NewClientWithConfig(config)
}

func decodeJSONLocation(jsonString string) (float64, float64, error) {
	type Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	var location Location
	err := json.Unmarshal([]byte(jsonString), &location)
	if err != nil {
		return 0, 0, fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	return location.Latitude, location.Longitude, nil
}

const locationRequest = `
Преобразуй местоположение в Москве в координаты. Если дано несколько местоположений, то возьми среднее из них.
Ответ дай в формате json(без указания json) c ключами latitude и longitude. Местоположение: %s`

func TransformLocation(location string) (float64, float64, error) {
	ans, err := request(fmt.Sprintf(locationRequest, location))
	if err != nil {
		return 0, 0, err
	}
	return decodeJSONLocation(ans)
}

func request(content string) (string, error) {

	resp, err := Connection.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: content,
				},
			},
			MaxCompletionTokens: maxTokens,
			Temperature:         temperature,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}

func (connection Client) Test() {

	resp, err := connection.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4oMini,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello 2+5=?",
				},
			},
			MaxCompletionTokens: maxTokens,
			Temperature:         temperature,
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}
