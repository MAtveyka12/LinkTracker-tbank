package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	gh "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/github"
	prod "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/producer"
	sof "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/domain/stackoverflow"
	clients "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/client"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/producer"
)

type Scheduler interface {
	AddLink(link string, userID, linkID int) error
	MonitorRepo(ctx context.Context, owner, repo, link string, userID, linkID int) error
	HandleCommits(ctx context.Context, ghClient clients.Client, owner, repo, link string, userID, linkID int) error
	HandlePullRequests(ctx context.Context, ghClient clients.Client, owner, repo, link string, userID, linkID int) error
	MonitorStackOverflow(ctx context.Context, questionID int, question, link string, userID, linkID int) error
	RemoveLink(link string) error
	Stop() error
}

type scheduler struct {
	logger          *slog.Logger
	clients         []clients.Client
	latestCommitSHA *sync.Map
	latestPR        *sync.Map
	latestSOAnswer  *sync.Map
	cancelFuncs     *sync.Map
	updateChan      chan prod.LinkUpdate
	commitChan      chan gh.Commit
	prChan          chan gh.PullRequest
	answerChan      chan sof.Answer
	producer        producer.ProdKafkaOrHTTP
	cfg             *config.Config
}

const (
	kafka = "kafka"
	http  = "http"
)

func NewScheduler(
	client []clients.Client,
	logger *slog.Logger,
	updateChan chan prod.LinkUpdate,
	httpProducer producer.ProdKafkaOrHTTP,
	cfg *config.Config,
) Scheduler {
	return &scheduler{
		logger:          logger,
		clients:         client,
		latestCommitSHA: &sync.Map{},
		latestPR:        &sync.Map{},
		latestSOAnswer:  &sync.Map{},
		cancelFuncs:     &sync.Map{},
		updateChan:      updateChan,
		commitChan:      make(chan gh.Commit, 10),
		prChan:          make(chan gh.PullRequest, 10),
		answerChan:      make(chan sof.Answer, 10),
		producer:        httpProducer,
		cfg:             cfg,
	}
}

func (s *scheduler) AddLink(link string, userID, linkID int) error {
	parts := strings.Split(link, "/")
	if len(parts) < 4 {
		return fmt.Errorf("incorrect link format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	s.cancelFuncs.Store(link, cancel)

	switch {
	case strings.Contains(link, "github.com"):
		owner, repo := parts[3], parts[4]
		go func() {
			err := s.MonitorRepo(ctx, owner, repo, link, userID, linkID)

			if err != nil {
				s.logger.Error("monitoring error GitHub")
			}
		}()
	case strings.Contains(link, "stackoverflow.com/questions/"):
		questionID, err := strconv.Atoi(parts[4])
		if err != nil {
			return fmt.Errorf("incorrect question ID format StackOverflow")
		}

		question := parts[5]

		go func() {
			err := s.MonitorStackOverflow(ctx, questionID, question, link, userID, linkID)

			if err != nil {
				s.logger.Error("monitoring error StackOverflow")
			}
		}()
	default:
		return fmt.Errorf("unsupported resource")
	}

	return nil
}

func (s *scheduler) getGitHubClient() clients.Client {
	for _, client := range s.clients {
		if client.GetType() == clients.GitHubType {
			return client
		}
	}

	return nil
}

func (s *scheduler) getStackOverflowClient() clients.Client {
	for _, client := range s.clients {
		if client.GetType() == clients.StackOverflowType {
			return client
		}
	}

	return nil
}

func (s *scheduler) MonitorRepo(ctx context.Context, owner, repo, link string, userID, linkID int) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	ghClient := s.getGitHubClient()
	if ghClient == nil {
		return fmt.Errorf("the GitHub client is not configured")
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Repository monitoring stopped", "link", link)
			return nil
		case <-ticker.C:
			if err := s.HandleCommits(ctx, ghClient, owner, repo, link, userID, linkID); err != nil {
				s.logger.Error("Processing error Commits", "error", err)
			}

			if err := s.HandlePullRequests(ctx, ghClient, owner, repo, link, userID, linkID); err != nil {
				s.logger.Error("Processing error PullRequests", "error", err)
			}
		}
	}
}

func (s *scheduler) HandleCommits(ctx context.Context, ghClient clients.Client, owner, repo, link string, userID, linkID int) error {
	commits, err := ghClient.GetLatestCommits(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("error when receiving Commits: %s", err.Error())
	}

	if len(commits) == 0 {
		return fmt.Errorf("error Commits were not found for a link with an ID: %d", linkID)
	}

	latestCommit := commits[0]
	val, ok := s.latestCommitSHA.Load(link)

	if !ok || latestCommit != val {
		s.latestCommitSHA.Store(link, latestCommit)
		s.commitChan <- latestCommit

		update := prod.LinkUpdate{
			ID:          int64(linkID),
			URL:         link,
			Description: latestCommit.Commit.Message,
			UserID:      int64(userID),
			Type:        "Commit",
		}

		switch s.cfg.MessageTransport {
		case kafka:
			s.updateChan <- update
		case http:
			err = s.producer.SendUpdate(update)

			if err != nil {
				s.logger.Error("Error sending an update to the bot with HTTP Producer", "error", err)
			}
		default:
			return fmt.Errorf("unknown message transport type: %s", s.cfg.MessageTransport)
		}
	}

	return nil
}

func (s *scheduler) HandlePullRequests(ctx context.Context, ghClient clients.Client, owner, repo, link string, userID, linkID int) error {
	pr, err := ghClient.GetLatestPRs(ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("error when receiving Pull Requests: %s", err.Error())
	}

	if len(pr) == 0 {
		return fmt.Errorf("error Pull Requests were not found for the link with the ID: %d", linkID)
	}

	latestPR := pr[0]
	if latestPR.ID == 0 || latestPR.Title == "" || latestPR.CreatedAt.IsZero() {
		s.logger.Info("PullRequests empty")
	}

	vall, ok := s.latestPR.Load(link)

	if !ok || latestPR != vall {
		s.latestPR.Store(link, latestPR)
		s.prChan <- latestPR

		update := prod.LinkUpdate{
			ID:  int64(linkID),
			URL: link,
			Description: fmt.Sprintf("Новый PullRequest: %s\nАвтор: %s\nСоздан: %s\nОписание: %s",
				latestPR.Title,
				latestPR.User.Login,
				latestPR.CreatedAt,
				truncateDescription(latestPR.Body)),
			UserID: int64(userID),
			Type:   "PullRequest",
		}

		switch s.cfg.MessageTransport {
		case kafka:
			s.updateChan <- update
		case http:
			err = s.producer.SendUpdatePR(update)

			if err != nil {
				s.logger.Error("Error sending an update to the bot with HTTP Producer", "error", err)
			}
		default:
			return fmt.Errorf("unknown message transport type: %s", s.cfg.MessageTransport)
		}
	}

	return nil
}

func (s *scheduler) MonitorStackOverflow(ctx context.Context, questionID int, question, link string, userID, linkID int) error {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	soClient := s.getStackOverflowClient()
	if soClient == nil {
		return fmt.Errorf("the StackOverflow client is not configured")
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("StackOverflow monitoring stopped", "link", link)
			return nil
		case <-ticker.C:
			answers, err := soClient.GetLatestAnswers(ctx, questionID)
			if err != nil {
				s.logger.Error("Error receiving StackOverflow responses", "error", err)
				continue
			}

			if len(answers) == 0 {
				continue
			}

			latestAnswer := answers[0]
			val, ok := s.latestSOAnswer.Load(link)

			if !ok || latestAnswer != val {
				s.latestSOAnswer.Store(link, latestAnswer)
				s.answerChan <- latestAnswer

				update := prod.LinkUpdate{
					ID:  int64(linkID),
					URL: link,
					Description: fmt.Sprintf("Новый ответ на вопрос: %s\nАвтор: %s\nСоздан: %s\nОписание: %s",
						question,
						latestAnswer.Owner.DisplayName,
						latestAnswer.GetCreationTime().Format(time.RFC3339),
						truncateDescription(latestAnswer.Body)),
					UserID: int64(userID),
					Type:   "Answer",
				}

				switch s.cfg.MessageTransport {
				case kafka:
					s.updateChan <- update
				case http:
					err = s.producer.SendUpdate(update)

					if err != nil {
						s.logger.Error("Error sending an update to the bot with HTTP Producer", "error", err)
					}
				default:
					return fmt.Errorf("unknown message transport type: %s", s.cfg.MessageTransport)
				}
			}
		}
	}
}

func (s *scheduler) RemoveLink(link string) error {
	if cancelFunc, ok := s.cancelFuncs.Load(link); ok {
		cancelFunc.(context.CancelFunc)()
		s.cancelFuncs.Delete(link)
		s.latestCommitSHA.Delete(link)
		s.latestPR.Delete(link)
		s.latestSOAnswer.Delete(link)
		s.logger.Info("Link deleted", "link", link)
	}

	return nil
}

func (s *scheduler) Stop() error {
	s.cancelFuncs.Range(func(_, value interface{}) bool {
		value.(context.CancelFunc)()
		return true
	})
	s.logger.Info("The scheduler is stopped")

	return nil
}

func truncateDescription(desc string) string {
	if len(desc) > 200 {
		return desc[:200] + "..."
	}

	return desc
}
