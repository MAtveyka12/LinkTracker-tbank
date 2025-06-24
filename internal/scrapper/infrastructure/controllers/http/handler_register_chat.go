package http

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/controllers/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlerRegisterChat(c *gin.Context) {
	slog.Info("Starting chat registration",
		slog.String("method", c.Request.Method),
		slog.String("path", c.FullPath()),
		slog.String("client_ip", c.ClientIP()),
	)

	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		slog.Warn("Invalid user ID parameter",
			slog.String("param", c.Param("id")),
			slog.String("error", err.Error()),
		)

		c.JSON(http.StatusBadRequest, dto.APIErrorResponse{
			Description: "Incorrect request parameters",
			Code:        "400",
		})

		return
	}

	slog.Debug("Parsed request parameters",
		slog.Int("user_id", userID),
	)

	slog.Debug("Calling service layer",
		slog.Int("user_id", userID),
	)

	err = h.service.RegisterChat(c, userID)
	if err != nil {
		slog.Error("Failed to register chat",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)

		c.JSON(http.StatusInternalServerError, dto.APIErrorResponse{
			Description: "Internal server error",
			Code:        "500",
		})

		return
	}

	slog.Info("Chat registered successfully",
		slog.Int("user_id", userID),
	)

	c.JSON(http.StatusOK, gin.H{"Description": "The chat is registered"})
}
