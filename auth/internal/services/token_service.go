package services

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
)

type TokenService struct {
	TokenRepository *repository.TokenRepository
}

func NewTokenService(tokenRepository *repository.TokenRepository) *TokenService {
	return &TokenService{
		TokenRepository: tokenRepository,
	}
}

func (s *TokenService) GetToken(userID uint, tokenType model.TokenType) (string, error) {
	token, err := security.GenerateToken()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])

	err = s.TokenRepository.Create(&model.Token{
		Hash:      hashedToken,
		UserID:    userID,
		Type:      tokenType,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		CreatedAt: time.Now().Unix(),
	})
	return token, nil
}

func (s *TokenService) VerifyToken(rawToken string, tokenType model.TokenType) (bool, error) {
	hash := sha256.Sum256([]byte(rawToken))
	hashedToken := hex.EncodeToString(hash[:])

	token, err := s.TokenRepository.GetOnHash(hashedToken, tokenType)
	if err != nil {
		return false, err
	}

	if token.ExpiresAt < time.Now().Unix() {
		return false, nil
	}

	if token.UsedAt != 0 {
		return false, nil
	}

	return true, nil
}
