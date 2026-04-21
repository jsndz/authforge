package repository

import (
	"github.com/jsndz/authforge/internal/model"
	"gorm.io/gorm"
)

type OauthRepo struct {
	db *gorm.DB
}

func NewOauthRepo(db *gorm.DB) *OauthRepo {
	return &OauthRepo{
		db: db,
	}
}

func (r *OauthRepo) Create(client *model.OauthClient) error {
	return r.db.Create(client).Error
}

func (r *OauthRepo) Get(clientID string) (*model.OauthClient, error) {
	var client model.OauthClient
	if err := r.db.Where("client_id = ?", clientID).First(&client).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *OauthRepo) Update(client *model.OauthClient) error {
	return r.db.Save(client).Error
}

func (r *OauthRepo) Delete(clientID string) error {
	return r.db.Where("client_id = ?", clientID).Delete(&model.OauthClient{}).Error
}

func (r *OauthRepo) List() ([]model.OauthClient, error) {
	var clients []model.OauthClient
	if err := r.db.Find(&clients).Error; err != nil {
		return nil, err
	}
	return clients, nil
}
