package services

import (
	"errors"
	"log"
	"time"

	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
)

type UserService struct {
	userRepository *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		userRepository: repo,
	}
}

func (s *UserService) Register(username, email, password string) (*model.User, error) {

	exists, err := s.userRepository.EmailExists(email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("email already registered")
	}

	hash, err := security.HashPassword(password, security.DefaultParams)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		UserName: username,
		Email:    email,
		Password: hash,
	}

	err = s.userRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(email, password string) (*model.User, error) {

	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	ok, err := security.VerifyPassword(password, user.Password)
	if err != nil || !ok {
		log.Printf("Login failed for email %s: %v", email, err)
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLoginAt = &now

	err = s.userRepository.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeactivateUser(id uint) error {

	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return err
	}

	user.IsActive = false

	return s.userRepository.Update(user)
}
