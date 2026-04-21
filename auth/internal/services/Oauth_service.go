package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	UserID              uint
	ClientId            string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	RedirectUri         string
}

func NewOAuthService(oauthRepo *repository.OauthRepo, sessionService *SessionService, redis *redis.Client) *OauthService {
	return &OauthService{
		redis:          redis,
		oauthRepo:      oauthRepo,
		sessionService: sessionService,
	}
}

func (s *OauthService) AuthorizeClient(ctx context.Context, clientId, sessionId, redirectUri, scope, codeChallenge, codeChallengeMethod string) (string, error) {
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
	allowedScope := strings.Split(client.Scopes, " ")
	requestedScope := strings.Split(scope, " ")
	if !isSubset(requestedScope, allowedScope) {
		return "", fmt.Errorf("scope not permitted")
	}
	authCode, err := security.GenerateAuthCode()
	if err != nil {
		return "", err
	}

	oauthUser := &OAuthUser{
		UserID:              userId,
		ClientId:            clientId,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		RedirectUri:         redirectUri,
	}
	data, err := json.Marshal(oauthUser)
	if err != nil {
		return "", err
	}

	err = s.redis.Set(ctx, authCode, data, time.Minute*5).Err()
	if err != nil {
		return "", err
	}
	return authCode, nil
}

func (s *OauthService) Token(ctx context.Context, clientId, code, redirectUri, codeVerifier string) (string, string, string, error) {
	var user OAuthUser
	val, err := s.redis.Get(ctx, code).Result()
	if err == redis.Nil {
		return "", "", "", errors.New("invalid_or_expired_code")
	}
	if err != nil {
		return "", "", "", err
	}
	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		return "", "", "", err
	}
	if user.ClientId != clientId {
		return "", "", "", errors.New("invalid_client")
	}
	if user.RedirectUri != redirectUri {
		return "", "", "", errors.New("invalid_redirect_uri")
	}
	if user.CodeChallenge == "" || codeVerifier == "" {
		return "", "", "", errors.New("pkce_required")
	}
	err = security.VerifyCodeChallenge(codeVerifier, user.CodeChallenge, user.CodeChallengeMethod)
	if err != nil {
		return "", "", "", errors.New("invalid_code_verifier")
	}
	if err := s.redis.Del(ctx, code).Err(); err != nil {
		return "", "", "", err
	}

	access_token, refresh_token, session_id, err := s.sessionService.CreateSessionTokens(ctx, user.UserID, user.Scope)
	if err != nil {
		return "", "", "", err
	}

	return access_token, refresh_token, session_id, nil
}

func isSubset(requested, allowed []string) bool {
	m := make(map[string]bool)
	for _, s := range allowed {
		m[s] = true
	}
	for _, s := range requested {
		if !m[s] {
			return false
		}
	}
	return true
}
