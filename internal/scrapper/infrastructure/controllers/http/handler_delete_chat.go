package http

import (
	"net/http"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/controllers/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlerDeleteChat(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIErrorResponse{
			Description: "Incorrect request parameters",
			Code:        "400",
		})

		return
	}

	err = h.service.DeleteChat(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIErrorResponse{
			Description: "Internal server error",
			Code:        "500",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"description": "Chat successfully deleted"})
}
