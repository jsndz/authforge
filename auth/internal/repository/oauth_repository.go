package repository

import "gorm.io/gorm"

type OauthRepo struct {
	db *gorm.DB
}

func NewOauthRepo(db *gorm.DB) *OauthRepo {
	return &OauthRepo{
		db: db,
	}
}
