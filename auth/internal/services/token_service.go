package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
	"github.com/jsndz/authforge/pkg/util"
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

func (s *TokenService) VerifyToken(rawToken string, tokenType model.TokenType) (*model.Token, error) {
	hashedToken := util.HashTokenWithSha256(rawToken)
	token, err := s.TokenRepository.GetOnHash(hashedToken, tokenType)
	if err != nil {
		return nil, err
	}
	if token.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	if token.UsedAt != 0 {
		return nil, errors.New("token already used")
	}
	if err := s.MarkTokenAsUsed(token); err != nil {
		return nil, err
	}
	return token, nil
}

func (s *TokenService) MarkTokenAsUsed(token *model.Token) error {
	if token.ExpiresAt < time.Now().Unix() {
		return nil
	}
	token.UsedAt = time.Now().Unix()
	return s.TokenRepository.MarkAsUsed(token.ID)
}
