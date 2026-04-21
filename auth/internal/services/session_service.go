package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/jsndz/authforge/internal/security"
	"github.com/jsndz/authforge/pkg/util"
	"github.com/redis/go-redis/v9"
)

type SessionService struct {
	redis     *redis.Client
	jwtSecret string
}

func NewSessionService(secret string, redisClient *redis.Client) *SessionService {
	return &SessionService{
		jwtSecret: secret,
		redis:     redisClient,
	}
}

func (s *SessionService) CreateSessionTokens(ctx context.Context, UserId uint, scope string) (string, string, string, error) {
	accessToken, err := util.CreateJWT(UserId, scope, 15*time.Minute, s.jwtSecret)
	if err != nil {
		log.Printf("Error creating access token: %v", err)
		return "", "", "", err
	}
	refreshToken, err := security.GenerateToken()
	if err != nil {
		log.Printf("Error creating refresh token: %v", err)
		return "", "", "", err
	}
	session_id := security.GenerateSessionId()
	hashedRefreshToken := util.HashTokenWithSha256(refreshToken)
	s.redis.Set(ctx, fmt.Sprintf("refresh:%s", hashedRefreshToken), UserId, 7*24*time.Hour)
	s.redis.SAdd(ctx, fmt.Sprintf("user_session:%d", UserId), hashedRefreshToken)
	s.redis.Set(ctx, fmt.Sprintf("session:%s", session_id), UserId, 7*24*time.Hour)
	return accessToken, refreshToken, session_id, err
}

func (s *SessionService) ValidateRefreshToken(ctx context.Context, refreshToken string) (uint, error) {

	hash := sha256.Sum256([]byte(refreshToken))
	key := "refresh:" + hex.EncodeToString(hash[:])

	userID, err := s.redis.Get(ctx, key).Uint64()
	if err == redis.Nil {
		return 0, fmt.Errorf("invalid or reused token")
	}
	if err != nil {
		return 0, fmt.Errorf("invalid refresh token")
	}

	return uint(userID), nil
}
func (s *SessionService) ValidateSession(ctx context.Context, sessionId string) (uint, error) {
	if sessionId == "" {
		return 0, fmt.Errorf("missing session")
	}

	userID, err := s.redis.Get(ctx, "session:"+sessionId).Uint64()
	if err == redis.Nil {
		return 0, fmt.Errorf("invalid session")
	}
	if err != nil {
		return 0, err
	}

	return uint(userID), nil
}
func (s *SessionService) RevokeToken(ctx context.Context, token string) error {
	s.redis.Del(ctx, fmt.Sprintf("refresh:%s", token))
	return nil
}

func (s *SessionService) RevokeSession(ctx context.Context, sessionId string) error {
	s.redis.Del(ctx, fmt.Sprintf("session:%s", sessionId))
	return nil
}
func (s *SessionService) BlacklistToken(ctx context.Context, token string, duration time.Duration) error {
	s.redis.Set(ctx, fmt.Sprintf("blacklist:%s", token), "blacklisted", duration)
	return nil
}

func (s *SessionService) AllSessionLogout(ctx context.Context, userID uint) error {
	sessionKeys, err := s.redis.SMembers(ctx, fmt.Sprintf("user_session:%d", userID)).Result()
	if err != nil {
		return err
	}

	for _, key := range sessionKeys {
		s.redis.Del(ctx, fmt.Sprintf("refresh:%s", key))
	}
	s.redis.Del(ctx, fmt.Sprintf("user_session:%d", userID))
	return nil
}

func (s *SessionService) DeleteToken(ctx context.Context, token string) error {
	hashedToken := util.HashTokenWithSha256(token)
	userID, err := s.redis.Get(ctx, "refresh:"+hashedToken).Uint64()
	if err != nil {
		return err
	}
	s.redis.Del(ctx, fmt.Sprintf("refresh:%s", hashedToken))
	s.redis.SRem(ctx, fmt.Sprintf("user_session:%d", userID), hashedToken)

	return nil
}
