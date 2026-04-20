package services

import "github.com/jsndz/authforge/internal/repository"

type OauthService struct {
	oauthRepo *repository.OauthRepo
}

func NewOAuthService(OauthRepo *repository.OauthRepo) *OauthService {
	return &OauthService{
		oauthRepo: OauthRepo,
	}
}
