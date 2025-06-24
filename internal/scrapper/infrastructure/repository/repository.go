package repository

import (
	"errors"
)

var (
	ErrNoRows        = errors.New("no rows found")
	ErrAlreadyExists = errors.New("already exists")
)

type Repository interface {
	LinkRepository
	ChatRepository
}

type repository struct {
	LinkRepository
	ChatRepository
}

func NewRepository(linkRepo LinkRepository, chatRepo ChatRepository) Repository {
	return &repository{
		LinkRepository: linkRepo,
		ChatRepository: chatRepo,
	}
}
