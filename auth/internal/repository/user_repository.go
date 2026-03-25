package repository

import (
	"time"

	"github.com/jsndz/authforge/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(DB *gorm.DB) *UserRepository {
	return &UserRepository{
		db: DB,
	}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var user model.User

	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var user model.User

	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User

	err := r.db.Where("user_name = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Update(userId uint, data map[string]interface{}) (*model.User, error) {
	var user model.User

	err := r.db.Model(&model.User{}).
		Clauses(clause.Returning{}).
		Where("id = ?", userId).
		Updates(data).
		Scan(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	var count int64

	err := r.db.Model(&model.User{}).
		Where("email = ?", email).
		Count(&count).Error

	return count > 0, err
}

func (r *UserRepository) UpdateVerification(verified bool, userId uint) (*model.User, error) {
	var user model.User

	err := r.db.Model(&model.User{}).
		Where("id = ?", userId).
		Updates(map[string]interface{}{
			"email_verified": verified,
			"updated_at":     time.Now(),
		}).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
