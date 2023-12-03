package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayloadFromEvent(t *testing.T) {

	type args struct {
		eventName string
		event     string
	}
	tests := []struct {
		name     string
		args     args
		expected payload
		wantErr  bool
	}{
		{
			name: "discussion created",
			args: args{
				eventName: "discussion",
				event: `{
					"action": "created",
					"discussion": {
						"body": "Some body",
						"title": "Some title",
						"node_id": "D_kwDOKgkPac4AWfor"
					},
					"user": {
						"login": "test"
					}
				}`,
			},
			expected: payload{title: "Some title", body: "Some body", nodeID: "D_kwDOKgkPac4AWfor", userLogin: "test", parentID: "D_kwDOKgkPac4AWfor"},
		},
		{
			name: "issue opened",
			args: args{
				eventName: "issues",
				event: `{
					"action": "opened",
					"issue": {
						"body": "Some body",
						"title": "Some title",
						"node_id": "D_kwDOKgkPac4AWfor"
					},
					"user": {
						"login": "test"
					}
				}`,
			},
			expected: payload{title: "Some title", body: "Some body", nodeID: "D_kwDOKgkPac4AWfor", userLogin: "test", parentID: "D_kwDOKgkPac4AWfor"},
		},
		{
			name: "pull_request opened",
			args: args{
				eventName: "pull_request",
				event: `{
					"action": "opened",
					"pull_request": {
						"body": "Some body",
						"title": "Some title",
						"node_id": "D_kwDOKgkPac4AWfor"
					},
					"user": {
						"login": "test"
					}
				}`,
			},
			expected: payload{title: "Some title", body: "Some body", nodeID: "D_kwDOKgkPac4AWfor", userLogin: "test", parentID: "D_kwDOKgkPac4AWfor"},
		},
		{
			name: "pull_request opened with empty body",
			args: args{
				eventName: "pull_request",
				event: `{
					"action": "opened",
					"pull_request": {
						"title": "Some title",
						"node_id": "D_kwDOKgkPac4AWfor"
					},
					"user": {
						"login": "test"
					}
				}`,
			},
			expected: payload{title: "Some title", nodeID: "D_kwDOKgkPac4AWfor", userLogin: "test", parentID: "D_kwDOKgkPac4AWfor"},
		},
		{
			name: "discussion comment created",
			args: args{
				eventName: "discussion_comment",
				event: `{
					"action": "created",
					"comment": {
						"body": "Some body",
						"node_id": "D_kwDOKgkPac4AWfor"
					},
					"discussion": {
						"node_id": "some_id"
					},
					"user": {
						"login": "test"
					}
				}`,
			},
			expected: payload{body: "Some body", nodeID: "D_kwDOKgkPac4AWfor", parentID: "some_id", userLogin: "test"},
		},
		{
			name: "issue comment created",
			args: args{
				eventName: "issue_comment",
				event: `{
					"action": "created",
					"comment": {
						"body": "Some body",
						"node_id": "D_kwDOKgkPac4AWfor"
					},
					"issue": {
						"node_id": "some_id"
					},
					"user": {
						"login": "test"
					}
				}`,
			},
			expected: payload{body: "Some body", nodeID: "D_kwDOKgkPac4AWfor", parentID: "some_id", userLogin: "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := payloadFromEvent(tt.args.eventName, strings.NewReader(tt.args.event))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, p)
		})
	}
}