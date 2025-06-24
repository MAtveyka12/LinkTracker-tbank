package http

import (
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/service"
)

type Link struct {
	URL    string `json:"link"`
	UserID int    `json:"user_id"`
	LinkID int    `json:"link_id"`
}

type Handler struct {
	service service.Service
}

func NewHandler(serv service.Service) *Handler {
	return &Handler{
		service: serv,
	}
}
