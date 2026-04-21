package model

type OauthClient struct {
	ID          uint   `gorm:"primaryKey"`
	ClientID    string `gorm:"unique;not null"`
	Secret      string
	RedirectUri string `gorm:"not null"`
	Scopes      string `gorm:"not null"`
}
