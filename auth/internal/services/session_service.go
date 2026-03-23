package services

import (
	"time"

	"github.com/jsndz/authforge/internal/security"
)

type SessionService struct {
	jwtSecret string
}

func NewSessionService(secret string) *SessionService {
	return &SessionService{
		jwtSecret: secret,
	}
}

func (s *SessionService) CreateSessionTokens(UserId uint) (string, string, error) {
	accessToken, err := security.CreateJWT(UserId, 15*time.Minute, s.jwtSecret)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := security.CreateJWT(UserId, 7*24*time.Hour, s.jwtSecret)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, err
}
