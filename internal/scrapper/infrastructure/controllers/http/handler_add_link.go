package http

import (
	"net/http"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/controllers/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlerAddLink(c *gin.Context) {
	var AddLinkRequest dto.AddLinkRequestDTO

	err := c.ShouldBindJSON(&AddLinkRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIErrorResponse{
			Description: "Incorrect request parameters",
			Code:        "400",
		})

		return
	}

	userID, err := strconv.Atoi(c.GetHeader("Tg-Chat-ID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIErrorResponse{
			Description: "Incorrect request parameters",
			Code:        "400",
		})

		return
	}

	linkID, err := h.service.AddLink(c.Request.Context(), userID, AddLinkRequest.Link)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIErrorResponse{
			Description: "Internal server error",
			Code:        "500",
		})

		return
	}

	c.JSON(http.StatusOK, dto.LinkResponseDTO{
		ID:      int64(linkID),
		URL:     AddLinkRequest.Link,
		Tags:    AddLinkRequest.Tags,
		Filters: AddLinkRequest.Filters,
	})
}
