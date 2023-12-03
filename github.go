package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GitHubGraphQLClient struct {
	token    string
	endpoint string
	client   *http.Client
}

func NewGithubClient(token string, endpoint string, client *http.Client) *GitHubGraphQLClient {
	if client == nil {
		client = &http.Client{}
	}

	return &GitHubGraphQLClient{
		token:    token,
		endpoint: endpoint,
		client:   client,
	}
}

type request struct {
	Query string `json:"query"`
}

func buildAddCommentRequest(subjectID string, body string) (string, error) {
	req := request{
		Query: fmt.Sprintf(`mutation{addComment(input:{subjectId:"%s", body:%s}){clientMutationId}}`, subjectID, body),
	}
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *GitHubGraphQLClient) AddComment(ctx context.Context, subjectID string, body string) error {
	payload, err := buildAddCommentRequest(subjectID, body)
	if err != nil {
		return fmt.Errorf("failed to build `addComment` request: %w", err)
	}
	if _, err := c.request(ctx, payload); err != nil {
		return fmt.Errorf("failed to add comment: %w", err)
	}
	return nil
}

func buildAddDiscussionCommentRequest(discussionID string, body string) (string, error) {
	req := request{
		Query: fmt.Sprintf(`mutation{addDiscussionComment(input:{discussionId:"%s", body:%s}){clientMutationId}}`, discussionID, body),
	}
	b, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (c *GitHubGraphQLClient) AddDiscussionComment(ctx context.Context, discussionID string, body string) error {
	payload, err := buildAddDiscussionCommentRequest(discussionID, body)
	if err != nil {
		return fmt.Errorf("failed to build `addDiscussionComment` request: %w", err)
	}
	if _, err := c.request(ctx, payload); err != nil {
		return fmt.Errorf("failed to add discussion comment: %w", err)
	}
	return nil
}

func (c *GitHubGraphQLClient) request(ctx context.Context, payload string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("bearer %s", c.token))
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read body: %w", err)
		}
		return nil, fmt.Errorf("status code: %d, body: %s", resp.StatusCode, string(b))
	}

	var r ghResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if r.HasErrors() {
		return nil, fmt.Errorf("bad request: %w", r)
	}

	return r.Data, err
}

type ghResponse struct {
	Data   json.RawMessage   `json:"data"`
	Errors []json.RawMessage `json:"errors"`
}

func (r ghResponse) HasErrors() bool {
	return len(r.Errors) > 0
}

func (r ghResponse) Error() string {
	if len(r.Errors) == 0 {
		return ""
	}
	var errs []string
	for _, e := range r.Errors {
		errs = append(errs, string(e))
	}

	return strings.Join(errs, ",")
}
