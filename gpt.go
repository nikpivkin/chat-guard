package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sashabaranov/go-openai"
)

type sentimentAssistant struct {
	model  string
	client *openai.Client
}

func newSentimentAssistant(token string, model string, httpClient *http.Client) *sentimentAssistant {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	clientCfg := openai.DefaultConfig(token)
	clientCfg.HTTPClient = httpClient
	client := openai.NewClientWithConfig(clientCfg)
	return &sentimentAssistant{client: client, model: model}
}

type sentimentAnalysisRequest struct {
	ty      string
	title   string
	content string
}

type sentimentAnalysisResponse struct {
	Sentiment   string `json:"sentiment"`
	Type        string `json:"type"`
	Explanation string `json:"explanation"`
}

var ErrEmptyMessage = errors.New("empty message")

func (a sentimentAssistant) Analyze(ctx context.Context, req sentimentAnalysisRequest) (sentimentAnalysisResponse, error) {
	var resp sentimentAnalysisResponse
	prompt := buildPrompt(req)
	chatResponse, err := a.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    a.model,
			Messages: prompt,
		},
	)
	if err != nil {
		return resp, fmt.Errorf("failed to create completion: %w", err)
	}

	msg := chatResponse.Choices[0].Message.Content
	if msg == "" {
		return resp, ErrEmptyMessage
	}

	if err := json.Unmarshal([]byte(msg), &resp); err != nil {
		return resp, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return resp, nil
}

func buildPrompt(req sentimentAnalysisRequest) []openai.ChatCompletionMessage {
	return []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: buildSystemPrompt(),
		},
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: buildUserPropmpt(req),
		},
	}
}

func buildUserPropmpt(req sentimentAnalysisRequest) string {
	return fmt.Sprintf("Type:%s\nTitle:%s\nContent:%s\n", req.ty, req.title, req.content)
}

func buildSystemPrompt() string {
	return `You're the moderator of the repository on github.

Your job is to analyse the titles and content of discussions, issues and comments and evaluate their sentiment. Give a negative rating for spam, legal adverts and insults.

Return the answer in json format. Example response:
{
  "sentiment": "negative",
  "type": "spam",
  "explanation": "The analyzed content exhibits characteristics associated with spam, including excessive use of promotional language and repeated, irrelevant information. This classification is based on linguistic patterns and content structure."
}

The sentiment field can contain one of two values: negative and positive. If the content has a positive sentiment, leave the type and explanation fields blank.
`
}
