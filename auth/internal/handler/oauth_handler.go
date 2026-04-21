package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/services"
)

type OauthHandler struct {
	oauthService *services.OauthService
}

type TokenRequest struct {
	ClientId string `json:"client_id" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

func NewOauthHandler(OauthService *services.OauthService) *OauthHandler {
	return &OauthHandler{
		oauthService: OauthService,
	}
}

func (h *OauthHandler) Authorize(c *gin.Context) {
	clientID := c.Query("client_id")
	redirectUri := c.Query("redirect_uri")
	scopes := c.Query("scopes")

	sessionId, err := c.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}
	authCode, err := h.oauthService.AuthorizeClient(c, clientID, sessionId, redirectUri, scopes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(302, redirectUri+"?code="+authCode)
}

func (h *OauthHandler) Token(c *gin.Context) {
	var req TokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	access_token, refresh_token, session_id, err := h.oauthService.Token(c, req.ClientId, req.Code)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.SetCookie("refresh_token", refresh_token, 7*24*3600, "/", "", true, true)
	c.SetCookie("session_id", session_id, 7*24*3600, "/", "", true, true)

	c.JSON(http.StatusCreated, gin.H{
		"access_token": access_token,
	})
}
