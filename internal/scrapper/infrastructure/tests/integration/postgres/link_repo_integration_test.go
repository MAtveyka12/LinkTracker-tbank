package postgres_test

import (
	"context"
	"log"
	"time"
)

func (s *TestSuite) TestAddLink() {
	startTime := time.Now()

	log.Printf("=== Starting TestAddLink ===")

	defer func() {
		log.Printf("=== TestAddLink completed in %v ===\n", time.Since(startTime))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chatID := 10
	linkURL := "https://example.com"

	log.Printf("Testing with chat ID: %d and URL: %s", chatID, linkURL)

	var exists bool

	log.Println("Checking initial chat existence...")

	err := s.pgPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM chats WHERE id = $1)", chatID).Scan(&exists)
	s.Require().NoError(err)

	if !exists {
		log.Println("Chat doesn't exist initially - as expected")
	} else {
		log.Println("Chat exists initially - unexpected")
	}

	s.False(exists, "Chat should not exist before registration")

	log.Println("Registering new chat...")

	err = s.chatRepo.RegisterChat(ctx, chatID)
	s.Require().NoError(err)

	log.Println("Chat registered successfully")
	log.Println("Verifying chat registration...")

	err = s.pgPool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM chats WHERE id = $1)", chatID).Scan(&exists)
	s.Require().NoError(err)

	if exists {
		log.Println("Chat exists after registration - as expected")
	} else {
		log.Println("Chat doesn't exist after registration - unexpected")
	}

	s.True(exists, "Chat should exist after registration")

	log.Println("Adding new link...")

	linkID, err := s.linkRepo.AddLink(ctx, chatID, linkURL)
	s.Require().NoError(err)

	log.Printf("Link added with ID: %d", linkID)

	s.NotZero(linkID, "Link ID should not be zero")

	var savedURL string

	log.Println("Verifying saved link...")

	err = s.pgPool.QueryRow(ctx, "SELECT url FROM links WHERE id = $1", linkID).Scan(&savedURL)
	s.Require().NoError(err)

	if savedURL == linkURL {
		log.Println("Saved URL matches original - verification successful")
	} else {
		log.Printf("URL mismatch: saved '%s' vs original '%s'", savedURL, linkURL)
	}

	s.Equal(linkURL, savedURL, "Saved link URL mismatch")
}

func (s *TestSuite) TestCheckEmpty() {
	startTime := time.Now()

	log.Printf("=== Starting TestCheckEmpty ===")

	defer func() {
		log.Printf("=== TestCheckEmpty completed in %v ===\n", time.Since(startTime))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Checking empty links table...")

	var count int

	err := s.pgPool.QueryRow(ctx, "SELECT COUNT(*) FROM links").Scan(&count)
	s.Require().NoError(err)

	if count == 0 {
		log.Println("Links table is empty - as expected")
	} else {
		log.Printf("Links table contains %d records - unexpected", count)
	}

	s.Equal(0, count, "Expected no links in table")
}
