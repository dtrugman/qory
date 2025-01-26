package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dtrugman/qory/lib/message"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	errorEmptyResponse = errors.New("empty response")
)

type client struct {
	openaiClient *openai.Client
}

type Client interface {
	AvailableModels() ([]string, error)
	Query(model string, messages []message.Message) (string, error)
}

func NewClient(apiKey *string, baseURL *string) Client {
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
	}
}

func (c *client) parseError(raw error) error {
	var apierr *openai.Error
	if !errors.As(raw, &apierr) {
		return raw
	}

	var errobj struct {
		Error openai.ErrorObject `json:"error"`
	}
	if err := json.Unmarshal([]byte(apierr.JSON.RawJSON()), &errobj); err != nil {
		return raw
	}

	return fmt.Errorf("%s", errobj.Error.Message)
}

func (c *client) AvailableModels() ([]string, error) {
	ctx := context.Background()

	modelNames := make([]string, 0)

	pager := c.openaiClient.Models.ListAutoPaging(ctx)
	if pager.Err() != nil {
		return nil, c.parseError(pager.Err())
	}

	for pager.Next() {
		modelNames = append(modelNames, pager.Current().ID)
	}

	return modelNames, nil
}

func (c *client) translateMessage(m message.Message) openai.ChatCompletionMessageParamUnion {
	switch m.Role {
	case message.RoleUser:
		return openai.UserMessage(m.Content)
	case message.RoleSystem:
		return openai.SystemMessage(m.Content)
	case message.RoleAssistant:
		return openai.AssistantMessage(m.Content)
	default:
		panic("unknown role")
	}
}

func (c *client) Query(model string, messages []message.Message) (string, error) {
	ctx := context.Background()

	openAIMessages := make([]openai.ChatCompletionMessageParamUnion, 0)
	for _, message := range messages {
		openAIMessage := c.translateMessage(message)
		openAIMessages = append(openAIMessages, openAIMessage)
	}

	stream := c.openaiClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(openAIMessages),
		Model:    openai.F(model),
	})

	var aggregator strings.Builder

	for stream.Next() {
		event := stream.Current()
		if len(event.Choices) > 0 {
			choice := event.Choices[0]
			if choice.Delta.Content != "" {
				content := choice.Delta.Content
				aggregator.WriteString(content)
				fmt.Print(content)
			}
		}
	}

	if err := stream.Err(); err != nil {
		parsed := c.parseError(err)
		fmt.Printf("Provider error: %v\n", parsed)
		return "", err
	}

	fmt.Println("")

	return aggregator.String(), nil
}
