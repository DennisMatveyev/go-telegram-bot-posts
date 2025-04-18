package app

import (
	"context"
	"log/slog"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type OpenAIsummarizer struct {
	client *openai.Client
	log    *slog.Logger
}

func NewOpenAIsummarizer(apiKey string, log *slog.Logger) *OpenAIsummarizer {
	return &OpenAIsummarizer{
		client: openai.NewClient(apiKey),
		log:    log,
	}
}

func (s *OpenAIsummarizer) Summarize(ctx context.Context, text string) (string, error) {
	request := openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Summarize the following text in approximately 400-500 symbols: " + text,
			},
		},
		MaxTokens:   125,
		Temperature: 0.7,
		TopP:        1,
	}
	response, err := s.client.CreateChatCompletion(ctx, request)
	if err != nil {
		s.log.Error("failed to create chat completion", "error", err.Error())
		return "", ErrSummarizerFailed
	}
	sammary := strings.TrimSpace(response.Choices[0].Message.Content)

	return sammary, nil
}
