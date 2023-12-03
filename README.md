# ChatGuard

## Description

ChatGuard is a GitHub Action designed to leverage ChatGPT for sentiment analysis of comments and discussions within your repository. It enhances community management by identifying and handling content that includes spam, advertisements, or offensive language. ChatGPT provides valuable insights into why specific comments are flagged.

## Usage

1. Set up Workflow:
- Create a new workflow file (e.g., .github/workflows/chatguard.yml).
- Configure the workflow to trigger on events where sentiment analysis is necessary.
yaml

```yaml
name: ChatGuard

on:
  issues:
    types:
      - opened
  issue_comment:
    types:
      - created
      - edited
  pull_request_review_comment:
    types:
      - created
      - edited
```

2. Add ChatGuard Action:

Add the ChatGuard action to your workflow, specifying the necessary parameters.

```yaml
jobs:
  analyze_content:
    runs-on: ubuntu-latest

    steps:
    - name: Analyze Comments with ChatGPT
      uses: nikpivkin/chat-guard@v0
      with:
        gh-token: ${{ secrets.GITHUB_TOKEN }}
        openai-api-key: ${{ secrets.OPENAI_API_KEY }}
```

## Inputs

| Name | Description | Default |
|---|---|---|
| `gh-token` | GitHub personal access token. | |
| `openai-api-key` | API token for ChatGPT. | Required |
| `gpt-model` | The ChatGPT model used. See available models [here](https://github.com/sashabaranov/go-openai/blob/master/completion.go#L20). | "gpt-3.5-turbo" |
| `timeout` | Timeout in seconds. | 60 |


## License
This project is licensed under the [MIT License](/LICENSE).
