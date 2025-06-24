package clients

import (
	"context"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/github"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/stackoverflow"
)

type ClientType string

const (
	GitHubType        ClientType = "github"
	StackOverflowType ClientType = "stackoverflow"
)

type Client interface {
	GetType() ClientType
	GetLatestCommits(ctx context.Context, owner, repo string) ([]github.Commit, error)
	GetLatestPRs(ctx context.Context, owner, repo string) ([]github.PullRequest, error)
	GetLatestAnswers(ctx context.Context, questionID int) ([]stackoverflow.Answer, error)
	GetLatestComments(ctx context.Context, questionID int) ([]stackoverflow.Comment, error)
}
