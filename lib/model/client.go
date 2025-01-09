package model

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	errorEmptyResponse = errors.New("empty response")
)

type client struct {
	openaiClient *openai.Client
	model        string
}

type Client interface {
	Query(systemPrompt *string, userPrompt string)
}

func NewClient(apiKey *string, baseURL *string, model string) Client {
	var options []option.RequestOption

	if apiKey != nil {
		options = append(options, option.WithAPIKey(*apiKey))
	}

	if baseURL != nil {
		options = append(options, option.WithBaseURL(*baseURL))
	}

	openaiClient := openai.NewClient(options...)

	return &client{
		openaiClient: openaiClient,
		model:        model,
	}
}

func (c *client) Query(systemPrompt *string, userPrompt string) {
	ctx := context.Background()

	messages := make([]openai.ChatCompletionMessageParamUnion, 0)
	if systemPrompt != nil {
		messages = append(messages, openai.SystemMessage(*systemPrompt))
	}
	messages = append(messages, openai.UserMessage(userPrompt))

	stream := c.openaiClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		Model:    openai.F(c.model),
	})

	for stream.Next() {
		event := stream.Current()
		if len(event.Choices) > 0 {
			choice := event.Choices[0]
			if choice.Delta.Content != "" {
				fmt.Print(choice.Delta.Content)
			}
		}
	}

	if err := stream.Err(); err != nil {
		fmt.Printf("Error: %v", err)
	}

	fmt.Println("")
}
