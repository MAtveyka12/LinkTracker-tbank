package http

import (
	"net/http"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/controllers/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlerGetAllLinks(c *gin.Context) {
	userID, err := strconv.Atoi(c.GetHeader("Tg-Chat-ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIErrorResponse{
			Description: "Incorrect request parameters",
			Code:        "400",
		})

		return
	}

	links, err := h.service.GetAllLinks(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIErrorResponse{
			Description: "Internal server error",
			Code:        "500",
		})

		return
	}

	linksDTO := make([]dto.LinkDTO, 0)

	for _, link := range links {
		linksDTO = append(linksDTO, dto.MapLinkToDTO(link))
	}

	c.JSON(http.StatusOK, dto.ListLinksResponseDTO{
		Links: linksDTO,
		Size:  len(linksDTO),
	})
}
