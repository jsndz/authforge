package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/services"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(service *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: service,
	}
}

type RegisterRequest struct {
	UserName string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UpdateUsernameRequest struct {
	Username string `json:"username" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {

	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := h.UserService.Register(
		req.UserName,
		req.Email,
		req.Password,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       user.ID,
		"username": user.UserName,
		"email":    user.Email,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	ip := c.ClientIP()
	user, err := h.UserService.Login(c, req.Email, req.Password, ip)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.SetCookie("refresh_token", user.RefreshToken, 7*24*3600, "/", "", true, true)
	c.JSON(http.StatusCreated, gin.H{
		"access_token": user.AccessToken,
		"username":     user.User.Username,
		"email":        user.User.Email,
	})
}

func (h *UserHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token is required",
		})
		return
	}

	user, err := h.UserService.VerifyEmail(c, token, model.TokenEmailVerification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.SetCookie("refresh_token", user.RefreshToken, 7*24*3600, "/", "", true, true)
	c.JSON(http.StatusCreated, gin.H{
		"access_token": user.AccessToken,
		"username":     user.User.Username,
		"email":        user.User.Email,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {

	refreshToken, _ := c.Cookie("refresh_token")
	accessToken := c.GetHeader("Authorization")
	accessToken = accessToken[len("Bearer "):]

	if refreshToken == "" || accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "refresh token and access token are required",
		})
		return
	}
	err := h.UserService.Logout(c, refreshToken, accessToken)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refresh_token", "", -1, "/", "", true, true)

	c.JSON(200, gin.H{"message": "logged out"})
}

func (h *UserHandler) UpdateUsername(c *gin.Context) {
	var req UpdateUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDValue.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	updatedUser, err := h.UserService.UpdateUsername(userID, req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       updatedUser.ID,
		"username": updatedUser.UserName,
		"email":    updatedUser.Email,
	})
}
