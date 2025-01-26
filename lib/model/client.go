package model

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	Query(model string, systemPrompt *string, userPrompt string) (string, error)
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

func (c *client) Query(model string, systemPrompt *string, userPrompt string) (string, error) {
	ctx := context.Background()

	messages := make([]openai.ChatCompletionMessageParamUnion, 0)
	if systemPrompt != nil {
		messages = append(messages, openai.SystemMessage(*systemPrompt))
	}
	messages = append(messages, openai.UserMessage(userPrompt))

	var aggregator strings.Builder

	stream := c.openaiClient.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F(messages),
		Model:    openai.F(model),
	})

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
		fmt.Printf("Error: %v", parsed)
		return "", err
	}

	fmt.Println("")

	return aggregator.String(), nil
}
