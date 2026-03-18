package repository

import (
	"time"

	"github.com/jsndz/authforge/internal/model"
	"gorm.io/gorm"
)

type TokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(DB *gorm.DB) *TokenRepository {
	return &TokenRepository{
		db: DB,
	}
}

func (r *TokenRepository) Create(token *model.Token) error {

	return r.db.Create(token).Error
}

func (r *TokenRepository) GetOnUserId(userID uint, tokenType string) (*model.Token, error) {
	var t model.Token
	if err := r.db.Where("user_id = ? AND type = ?", userID, tokenType).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) GetOnHash(hash string, tokenType model.TokenType) (*model.Token, error) {
	var t model.Token
	if err := r.db.Where("hash = ? AND type = ?", hash, tokenType).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) Delete(id uint) (*model.Token, error) {
	var t model.Token
	if err := r.db.Delete(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TokenRepository) MarkAsUsed(id uint) error {
	return r.db.Model(&model.Token{}).Where("id = ?", id).Update("used_at", time.Now().Unix()).Error
}
