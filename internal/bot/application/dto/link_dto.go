package dto

import "github.com/es-debug/backend-academy-2024-go-template/internal/bot/domain"

type LinkDTO struct {
	URL    string `json:"link"`
	LinkID int64  `json:"link_id"`
	UserID int64  `json:"user_id"`
}

func FromDomain(link domain.Link) LinkDTO {
	return LinkDTO{
		URL:    link.URL,
		LinkID: link.LinkID,
		UserID: link.UserID,
	}
}

func ToDomain(dto LinkDTO) domain.Link {
	return domain.Link{
		URL:    dto.URL,
		LinkID: dto.LinkID,
		UserID: dto.UserID,
	}
}
