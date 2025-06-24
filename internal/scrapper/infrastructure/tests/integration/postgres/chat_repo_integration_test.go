package postgres_test

import (
	"context"
	"log"
	"time"
)

func (s *TestSuite) TestCreateChat() {
	startTime := time.Now()

	log.Printf("=== Starting TestCreateChat ===")

	defer func() {
		log.Printf("=== TestCreateChat completed in %v ===\n", time.Since(startTime))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	exampleUserID := 10

	log.Printf("Testing with user ID: %d", exampleUserID)
	log.Println("Registering new chat...")

	err := s.chatRepo.RegisterChat(ctx, exampleUserID)
	s.Require().NoError(err)

	log.Println("Chat registered successfully")

	var exists bool

	log.Println("Checking chat existence in database...")

	err = s.pgPool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)", exampleUserID).Scan(&exists)
	s.Require().NoError(err)

	if exists {
		log.Println("Chat found in database - verification successful")
	} else {
		log.Println("Chat not found in database - verification failed")
	}

	s.True(exists, "Chat should exist after creation")
}

func (s *TestSuite) TestDeleteChat() {
	startTime := time.Now()

	log.Printf("=== Starting TestDeleteChat ===")

	defer func() {
		log.Printf("=== TestDeleteChat completed in %v ===\n", time.Since(startTime))
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exampleUserID := 10

	log.Printf("Testing with user ID: %d", exampleUserID)
	log.Println("Registering chat for deletion test...")

	err := s.chatRepo.RegisterChat(ctx, exampleUserID)
	s.Require().NoError(err)

	log.Println("Chat registered successfully")

	var existsBefore bool

	log.Println("Checking pre-deletion chat existence...")

	err = s.pgPool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)", exampleUserID).Scan(&existsBefore)
	s.Require().NoError(err)

	if existsBefore {
		log.Println("Chat exists before deletion - as expected")
	} else {
		log.Println("Chat doesn't exist before deletion - unexpected")
	}

	s.True(existsBefore, "Chat should exist before deletion")

	log.Println("Deleting chat...")

	err = s.chatRepo.DeleteChat(ctx, exampleUserID)
	s.Require().NoError(err)

	log.Println("Chat deletion completed")

	var existsAfter bool

	log.Println("Checking post-deletion chat existence...")

	err = s.pgPool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chats WHERE id = $1)", exampleUserID).Scan(&existsAfter)
	s.Require().NoError(err)

	if !existsAfter {
		log.Println("Chat doesn't exist after deletion - as expected")
	} else {
		log.Println("Chat still exists after deletion - unexpected")
	}

	s.False(existsAfter, "Chat should not exist after deletion")
}
