package model

type OauthClient struct {
	ID          uint   `gorm:"primaryKey"`
	ClientId    uint   `gorm:"not null"`
	RedirectUri string `gorm:"not null"`
	Scopes      string `gorm:"not null"`
}
