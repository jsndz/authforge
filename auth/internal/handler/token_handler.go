package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/services"
)

type TokenHandler struct {
	tokenService *services.TokenService
}

func NewTokenHandler(service *services.TokenService) *TokenHandler {
	return &TokenHandler{
		tokenService: service,
	}
}

func (h *TokenHandler) VerifyEmailToken(c *gin.Context) {
	token := c.Query("token")
	ok, err := h.tokenService.VerifyToken(token, model.TokenEmailVerification)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}
