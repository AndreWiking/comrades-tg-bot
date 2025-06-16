package gpt

import (
	"ComradesTG/db"
	"ComradesTG/settings"
	"context"
	"encoding/json"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"log"
	"os"
)

const (
	BaseUrl     = "https://api.proxyapi.ru/openai/v1"
	maxTokens   = 1000
	temperature = 0.8
)

type Client struct {
	client *openai.Client
}

func NewClient() *Client {
	apiKey, ok := os.LookupEnv(settings.GptApiKeyName)
	if !ok {
		log.Fatalf("%s not found in environment variables\n", settings.GptApiKeyName)
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = BaseUrl
	return &Client{openai.NewClientWithConfig(config)}
}

func decodeJSONLocation(jsonString string) (float64, float64, error) {
	type Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	var location Location
	err := json.Unmarshal([]byte(jsonString), &location)
	if err != nil {
		return 0, 0, fmt.Errorf("error unmarshalling json: %w", err)
	}

	return location.Latitude, location.Longitude, nil
}

const locationRequest = `
Преобразуй местоположение в Москве в координаты. Если дано несколько местоположений, то возьми среднее из них.
Ответ дай в формате json(без указания json) c ключами latitude и longitude. Местоположение: %s`

func (c *Client) TransformLocation(location string) (float64, float64, error) {
	ans, err := c.request(fmt.Sprintf(locationRequest, location))
	if err != nil {
		return 0, 0, err
	}
	return decodeJSONLocation(ans)
}

const postTypeRequest = `
Определи тип объявление, верни true, если это объявление о поиске соседа для совместной аренды квартиры, false иначе.
Ответ дай в формате json(без указания json) c ключом type.
Текс объявления: %s`

func (c *Client) detectPostType(text string) (bool, error) {
	ans, err := c.request(fmt.Sprintf(postTypeRequest, text))
	if err != nil {
		return false, err
	}
	type postType struct {
		Type bool `json:"type"`
	}

	var pType postType
	if err := json.Unmarshal([]byte(ans), &pType); err != nil {
		return false, fmt.Errorf("error unmarshalling json: %w", err)
	}

	return pType.Type, nil
}

func (c *Client) DetectAllPostsType(connection *db.Connection, posts []db.PostVK) {
	for _, post := range posts {
		if post.Type != db.PostVkTypeNotSet {
			continue
		}
		postType, err := c.detectPostType(post.Text)
		if err != nil {
			log.Printf(fmt.Errorf("detect post type failed: %w", err).Error())
			continue
		}
		var resType db.PostVkType
		if postType {
			resType = db.PostVkTypeFindRoommate
		} else {
			resType = db.PostVkTypeOther
		}
		post.Type = resType
		if err := connection.SetVkPostType(post.Id, resType); err != nil {
			log.Printf(fmt.Errorf("set post type failed: %w", err).Error())
			continue
		}
	}
}

func (c *Client) request(content string) (string, error) {

	resp, err := c.client.CreateChatCompletion(
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

func (c *Client) GenerateSpamMessage(matches string, botUrl string) (string, error) {

	request := "Перепиши текст:\n" + settings.SpamMessagePattern

	if pattern, err := c.request(request); err != nil {
		return "", err
	} else {
		fmt.Println(pattern)
		return fmt.Sprintf(pattern, "\n"+matches, "\n"+botUrl), nil
	}

}

func (c *Client) Test() {

	resp, err := c.client.CreateChatCompletion(
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
