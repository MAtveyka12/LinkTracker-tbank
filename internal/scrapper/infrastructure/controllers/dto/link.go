package dto

import "github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/application/service"

type LinkDTO struct {
	URL string `json:"url" example:"https://github.com/epchamp001/avito-tech-merch"`
}

func MapLinkToDTO(link service.Link) LinkDTO {
	return LinkDTO{
		URL: link.URL,
	}
}

func MapLinkDTOToLink(link LinkDTO) service.Link {
	return service.Link{
		URL: link.URL,
	}
}
