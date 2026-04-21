package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
	"github.com/redis/go-redis/v9"
)

type OauthService struct {
	oauthRepo      *repository.OauthRepo
	sessionService *SessionService
	redis          *redis.Client
}

type OAuthUser struct {
	UserID   uint
	ClientId string
	Scope    string
}

func NewOAuthService(oauthRepo *repository.OauthRepo, sessionService *SessionService, redis *redis.Client) *OauthService {
	return &OauthService{
		redis:          redis,
		oauthRepo:      oauthRepo,
		sessionService: sessionService,
	}
}

func (s *OauthService) AuthorizeClient(ctx context.Context, clientId, sessionId, redirectUri, scope string) (string, error) {
	userId, err := s.sessionService.ValidateSession(ctx, sessionId)
	if err != nil {
		return "", err
	}
	client, err := s.oauthRepo.Get(clientId)
	if err != nil {
		if err.Error() == "client not found" {
			return "", fmt.Errorf("invalid client_id")
		}
		return "", err
	}
	if client.RedirectUri != redirectUri {
		return "", fmt.Errorf("invalid redirect URI")
	}
	authCode, err := security.GenerateAuthCode()
	if err != nil {
		return "", err
	}

	s.redis.Set(ctx, authCode, &OAuthUser{
		UserID:   userId,
		ClientId: clientId,
		Scope:    scope,
	}, time.Minute*5)
	return authCode, nil
}

func (s *OauthService) Token(ctx context.Context, clientId, code string) (string, string, string, error) {
	var user OAuthUser
	err := s.redis.Get(ctx, code).Scan(&user)
	if err == redis.Nil {
		return "", "", "", errors.New("invalid_or_expired_code")
	}
	if err != nil {
		return "", "", "", err
	}
	if user.ClientId != clientId {
		return "", "", "", errors.New("invalid_client")
	}

	if err := s.redis.Del(ctx, code).Err(); err != nil {
		return "", "", "", err
	}

	access_token, refresh_token, session_id, err := s.sessionService.CreateSessionTokens(ctx, user.UserID)
	if err != nil {
		return "", "", "", err
	}

	return access_token, refresh_token, session_id, nil
}
