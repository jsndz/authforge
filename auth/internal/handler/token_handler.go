package handler

import (
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
