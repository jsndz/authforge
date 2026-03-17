package model

type TokenType string

func (t TokenType) IsValid() bool {
	switch t {
	case TokenEmailVerification,
		TokenPasswordReset:
		return true
	}
	return false
}

const (
	TokenEmailVerification TokenType = "email_verification"
	TokenPasswordReset     TokenType = "password_reset"
)

type Token struct {
	ID        uint      `gorm:"primaryKey"`
	Hash      string    `gorm:"size:255;not null;index:idx_hash_type"`
	UserID    uint      `gorm:"index:idx_user_type"`
	Type      TokenType `gorm:"type:varchar(50);not null;index:idx_hash_type;index:idx_user_type"`
	ExpiresAt int64     `gorm:"not null;index"`
	CreatedAt int64     `gorm:"not null"`
	UsedAt    int64     `gorm:"default:null"`
}
