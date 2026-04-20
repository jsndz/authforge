package handler

import "github.com/jsndz/authforge/internal/services"

type OauthHandler struct {
	oauthService *services.OauthService
}

func NewOauthHandler(OauthService *services.OauthService) *OauthHandler {
	return &OauthHandler{
		oauthService: OauthService,
	}
}
