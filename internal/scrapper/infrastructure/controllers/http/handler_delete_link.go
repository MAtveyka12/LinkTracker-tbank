package http

import (
	"net/http"
	"strconv"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/controllers/dto"
	"github.com/gin-gonic/gin"
)

func (h *Handler) HandlerDeleteLink(c *gin.Context) {
	var deleteLinkReq dto.DeleteLinkRequestDTO

	err := c.ShouldBindJSON(&deleteLinkReq)
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

	err = h.service.DeleteLink(c.Request.Context(), userID, deleteLinkReq.Link)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMsg := "Error when deleting a link"

		if err.Error() == "the link was not found" {
			statusCode = http.StatusNotFound
			errorMsg = "the link was not found"
		}

		c.JSON(statusCode, dto.APIErrorResponse{
			Description: errorMsg,
			Code:        strconv.Itoa(statusCode),
		})

		return
	}

	c.JSON(http.StatusOK, dto.LinkResponseDTO{
		URL: deleteLinkReq.Link,
	})
}
