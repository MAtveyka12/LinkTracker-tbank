package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/application/dto"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"
	"github.com/es-debug/backend-academy-2024-go-template/internal/bot/infrastructure/scrapper"
	"github.com/es-debug/backend-academy-2024-go-template/pkg/reliability"
)

type ScraperHTTPClient struct {
	BaseURL string
	client  reliability.HTTPClient
}

type Link struct {
	URL string `json:"url"`
}

type ScraperResponse struct {
	Links []Link `json:"links"`
	Size  int    `json:"size"`
}

func NewScraperHTTPClient(
	baseURL string,
	config *reliability.RetryConfig,
	rateLimitConfig *reliability.RateLimitConfig,
	circuitBreakerConfig *reliability.CircuitBreakerConfig,
) scrapper.ScraperClient {
	client := reliability.NewHTTPClientWithReliability(
		10*time.Second,
		config,
		rateLimitConfig,
		circuitBreakerConfig,
	)

	return &ScraperHTTPClient{
		BaseURL: baseURL,
		client:  client,
	}
}

func (s *ScraperHTTPClient) GetLinks(ctx context.Context, userID int64) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/links", s.BaseURL), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error when creating the request: %s", err.Error())
	}

	req.Header.Set("Tg-Chat-ID", fmt.Sprintf("%d", userID))

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error when sending a scrapper request: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("the scrapper returned an error: %s, status: %d", string(body), resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("error reading the response from the scrapper: %s", err.Error())
	}

	var scraperResponse ScraperResponse

	if err := json.Unmarshal(body, &scraperResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %s", err.Error())
	}

	links := make([]string, 0)
	for _, link := range scraperResponse.Links {
		links = append(links, link.URL)
	}

	return links, nil
}

func (s *ScraperHTTPClient) RegisterUser(_ context.Context, user domain.User) error {
	url := fmt.Sprintf("%s/tg-chat/%d", s.BaseURL, user.TelegramID)
	userData, err := json.Marshal(user)

	if err != nil {
		return fmt.Errorf("error when encoding the user: %s", err.Error())
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(userData))
	if err != nil {
		return fmt.Errorf("error when creating the request: %s", err.Error())
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("error during user registration: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return fmt.Errorf("error reading the response from the scrapper: %s", err.Error())
		}

		return fmt.Errorf("the scrapper returned an error: %s, status: %d", string(body), resp.StatusCode)
	}

	return nil
}

func (s *ScraperHTTPClient) AddLink(_ context.Context, link domain.Link) error {
	linkDTO := dto.FromDomain(link)
	linkData, err := json.Marshal(linkDTO)

	if err != nil {
		return fmt.Errorf("error encoding link data in JSON: %s", err.Error())
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/links", s.BaseURL), bytes.NewBuffer(linkData))

	if err != nil {
		return fmt.Errorf("error when creating the request: %s", err.Error())
	}

	req.Header.Set("Tg-Chat-ID", fmt.Sprintf("%d", link.UserID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("error when sending a scrapper request: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return fmt.Errorf("error reading the response body: %s", err.Error())
		}

		return fmt.Errorf("the scrapper returned an error: %s, status: %d", string(body), resp.StatusCode)
	}

	return nil
}

func (s *ScraperHTTPClient) RemoveLink(_ context.Context, link string, userID int64) error {
	requestData := dto.LinkDTO{
		URL:    link,
		UserID: userID,
	}

	requestBody, err := json.Marshal(requestData)

	if err != nil {
		return fmt.Errorf("error encoding data in JSON: %s", err.Error())
	}

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/links", s.BaseURL), bytes.NewBuffer(requestBody))

	if err != nil {
		return err
	}

	req.Header.Set("Tg-Chat-ID", fmt.Sprintf("%d", userID))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)

	if err != nil {
		return fmt.Errorf("error when sending a scrapper request: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)

		if err != nil {
			return fmt.Errorf("error reading the response body: %s", err.Error())
		}

		return fmt.Errorf("the scrapper returned an error: %s, status: %d", string(body), resp.StatusCode)
	}

	return nil
}
