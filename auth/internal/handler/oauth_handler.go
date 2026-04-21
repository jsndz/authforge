package handler

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/services"
)

type OauthHandler struct {
	oauthService *services.OauthService
}

type TokenRequest struct {
	ClientId     string `form:"client_id" binding:"required"`
	Code         string `form:"code" binding:"required"`
	CodeVerifier string `form:"code_verifier" binding:"required"`
	RedirectUri  string `form:"redirect_uri" binding:"required"`
	GrantType    string `form:"grant_type" binding:"required"`
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
	state := c.Query("state")
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state_required"})
		return
	}
	sessionId, err := c.Cookie("session_id")
	code_challenge := c.Query("code_challenge")
	code_challenge_method := c.Query("code_challenge_method")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session token"})
		return
	}

	authCode, err := h.oauthService.AuthorizeClient(c, clientID, sessionId, redirectUri, scopes, code_challenge, code_challenge_method)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	u, _ := url.Parse(redirectUri)
	q := u.Query()
	q.Set("code", authCode)
	q.Set("state", state)
	u.RawQuery = q.Encode()

	c.Redirect(302, u.String())
}

func (h *OauthHandler) Token(c *gin.Context) {
	var req TokenRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	if req.GrantType != "authorization_code" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported_grant_type"})
		return
	}
	access_token, refresh_token, session_id, id_token, err := h.oauthService.Token(c, req.ClientId, req.Code, req.RedirectUri, req.CodeVerifier)

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
		"id_token":     id_token,
		"token_type":   "Bearer",
		"expires_in":   900,
	})
}
