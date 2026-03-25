package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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

func (s *SessionService) CreateSessionTokens(ctx context.Context, UserId uint) (string, string, error) {
	accessToken, err := util.CreateJWT(UserId, 15*time.Minute, s.jwtSecret)
	if err != nil {
		return "", "", err
	}
	refreshToken, err := security.GenerateToken()
	if err != nil {
		return "", "", err
	}
	hashedRefreshToken := util.HashTokenWithSha256(refreshToken)
	s.redis.Set(ctx, fmt.Sprintf("refresh:%s", hashedRefreshToken), UserId, 7*24*time.Hour)
	s.redis.SAdd(ctx, fmt.Sprintf("user_session:%d", UserId), hashedRefreshToken)
	return accessToken, refreshToken, err
}

func (s *SessionService) ValidateRefreshToken(ctx context.Context, refreshToken string) (uint, error) {

	hash := sha256.Sum256([]byte(refreshToken))
	key := "refresh:" + hex.EncodeToString(hash[:])

	userID, err := s.redis.Get(ctx, key).Uint64()
	if err != nil {
		return 0, fmt.Errorf("invalid refresh token")
	}

	return uint(userID), nil
}

func (s *SessionService) RevokeToken(ctx context.Context, token string) error {
	s.redis.Del(ctx, fmt.Sprintf("refresh:%s", token))
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
