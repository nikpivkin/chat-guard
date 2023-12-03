package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

const defaultTimeoutS = 60

func envOrFatal(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("env %q is required\n", key)
	}
	return val
}

func main() {
	log.Println("Starting the ChatGuard Action.")
	timeout := flag.Int("timeout", defaultTimeoutS, fmt.Sprintf("timeout in seconds (default %ds)", defaultTimeoutS))
	gptModel := flag.String("gpt-model", openai.GPT3Dot5Turbo, fmt.Sprintf("the chat-gpt model used (default %s)", openai.GPT3Dot5Turbo))

	flag.Parse()

	ghRepo := envOrFatal("GITHUB_REPOSITORY")
	parts := strings.Split(ghRepo, "/")

	c := config{
		timeout:         *timeout,
		eventName:       envOrFatal("GITHUB_EVENT_NAME"),
		eventPath:       envOrFatal("GITHUB_EVENT_PATH"),
		gptToken:        envOrFatal("OPENAI_API_KEY"),
		gptModel:        *gptModel,
		ghToken:         envOrFatal("GITHUB_TOKEN"),
		graphQLEndpoint: envOrFatal("GITHUB_GRAPHQL_URL"),
		repoOwner:       parts[0],
		repoName:        parts[1],
	}

	if err := run(c); err != nil {
		log.Fatal(err)
	}
}

type config struct {
	timeout         int
	eventName       string
	eventPath       string
	gptToken        string
	gptModel        string
	ghToken         string
	graphQLEndpoint string
	repoOwner       string
	repoName        string
}

func run(cfg config) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.timeout)*time.Second)
	defer cancel()

	ef, err := os.Open(cfg.eventPath)
	if err != nil {
		return fmt.Errorf("failed to open event file: %w", err)
	}

	payload, err := payloadFromEvent(cfg.eventName, ef)
	if err != nil {
		return fmt.Errorf("failed to parse event: %w", err)
	}

	assistant := newSentimentAssistant(cfg.gptToken, cfg.gptModel, nil)

	analysisResult, err := assistant.Analyze(ctx, sentimentAnalysisRequest{
		ty:      cfg.eventName,
		title:   payload.title,
		content: payload.body,
	})

	if err != nil {
		return err
	}

	if analysisResult.Sentiment == "positive" {
		log.Println("Content has a positive sentiment")
		return nil
	}

	ghapi := NewGithubClient(cfg.ghToken, cfg.graphQLEndpoint, nil)

	var addCommentFn = ghapi.AddComment
	switch cfg.eventName {
	case "discussion", "discussion_comment":
		addCommentFn = ghapi.AddDiscussionComment
	}

	comment := createComment(eventNameToArtifactName(cfg.eventName), analysisResult.Type, analysisResult.Explanation, payload.userLogin)
	if err := addCommentFn(ctx, payload.parentID, comment); err != nil {
		return err
	}

	return nil
}

func createComment(artifactName string, ty string, explanation string, user string) string {
	body := `🛡️ ChatGuard Analysis: The content has been reviewed, and based on sentiment analysis, it has been identified as %s. 
	
Explanation: %s

This %s has been labeled accordingly for further review. 

@%s please ensure future contributions align with our community guidelines. Thank you! 🚀`

	body = fmt.Sprintf(body, ty, explanation, artifactName, user)

	footer := "\n\n*Note: This message is generated automatically, and the labels were assigned based on the analysis of the %s's content.*"

	footer = fmt.Sprintf(footer, artifactName)
	return body + footer
}

type payload struct {
	// Id of the object to which the artefact belongs. For the discussion comment is the discussion, for the discussion itself.
	parentID  string
	nodeID    string
	title     string
	body      string
	userLogin string
}

func payloadFromEvent(eventName string, r io.Reader) (payload, error) {
	var event map[string]any
	if err := json.NewDecoder(r).Decode(&event); err != nil {
		return payload{}, fmt.Errorf("failed to decode %s event: %w", eventName, err)
	}

	obj, exist := event[eventNameToArtifactName(eventName)]
	if !exist {
		return payload{}, fmt.Errorf("invalid event name %q", eventName)
	}

	p := payload{}
	if m, ok := obj.(map[string]any); ok {
		p.nodeID = m["node_id"].(string)
		if title, ok := m["title"].(string); ok {
			p.title = title
		}
		if body, ok := m["body"].(string); ok {
			p.body = body
		}
		p.parentID = p.nodeID

	}

	switch eventName {
	case "issue_comment":
		p.parentID = event["issue"].(map[string]any)["node_id"].(string)
	case "discussion_comment":
		p.parentID = event["discussion"].(map[string]any)["node_id"].(string)
	}

	if u, ok := event["user"].(map[string]any); ok {
		p.userLogin = u["login"].(string)
	}

	return p, nil
}

func eventNameToArtifactName(eventName string) string {
	artifactName := eventName
	switch eventName {
	case "issues":
		artifactName = "issue"
	case "discussion_comment", "issue_comment":
		artifactName = "comment"
	}

	return artifactName
}
