package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
	"github.com/jsndz/authforge/pkg/email"
)

type UserService struct {
	userRepository *repository.UserRepository
	tokenService   *TokenService
	sessionService *SessionService
	emailService   *email.EmailService
}

type UserResponse struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

func NewUserService(
	repo *repository.UserRepository,
	tokenService *TokenService,
	sessionService *SessionService,
	emailService *email.EmailService,
) *UserService {
	return &UserService{
		userRepository: repo,
		tokenService:   tokenService,
		sessionService: sessionService,
		emailService:   emailService,
	}
}

func (s *UserService) Register(username, useremail, password string) (*model.User, error) {

	if username == "" || useremail == "" || password == "" {
		return nil, errors.New("all fields are required")
	}

	exists, err := s.userRepository.EmailExists(useremail)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("email already registered")
	}
	if err := security.PasswordStrengthValidation(password); err != nil {
		return nil, err
	}
	hash, err := security.HashPassword(password, security.DefaultParams)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		UserName: username,
		Email:    useremail,
		Password: hash,
		IsActive: true,
	}

	err = s.userRepository.Create(user)
	if err != nil {
		return nil, err
	}

	token, err := s.tokenService.GetToken(user.ID, model.TokenEmailVerification)
	if err != nil {
		return nil, err
	}

	err = s.emailService.SendEmailVerification(user.Email, token)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, useremail, password string) (*LoginResponse, error) {

	user, err := s.userRepository.FindByEmail(useremail)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if !user.EmailVerified {
		return nil, errors.New("email not verified")
	}

	ok, err := security.VerifyPassword(password, user.Password)
	if err != nil || !ok {
		return nil, errors.New("invalid credentials")
	}

	now := time.Now()
	user.LastLoginAt = &now

	err = s.userRepository.Update(user)
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.sessionService.CreateSessionTokens(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		User: UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.UserName,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, RefreshToken string, accessToken string) error {
	s.sessionService.BlacklistToken(ctx, accessToken, 15*time.Minute)
	return s.sessionService.RevokeToken(ctx, RefreshToken)
}

func (s *UserService) DeactivateUser(id uint) error {

	user, err := s.userRepository.FindByID(id)
	if err != nil {
		return err
	}

	if !user.IsActive {
		return errors.New("user already deactivated")
	}

	user.IsActive = false

	return s.userRepository.Update(user)
}

func (s *UserService) VerifyEmail(ctx context.Context, rawToken string, tokenType model.TokenType) (LoginResponse, error) {

	token, err := s.tokenService.VerifyToken(rawToken, tokenType)
	if err != nil {
		return LoginResponse{}, err
	}

	user, err := s.userRepository.UpdateVerification(true, token.UserID)
	if err != nil {
		return LoginResponse{}, err
	}

	accessToken, refreshToken, err := s.sessionService.CreateSessionTokens(ctx, token.UserID)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		User: UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.UserName,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil

}

func (s *UserService) UpdateUsername(userID uint, username string) (*model.User, error) {
	newUsername := strings.TrimSpace(username)
	if newUsername == "" {
		return nil, errors.New("username is required")
	}

	user, err := s.userRepository.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if user.UserName == newUsername {
		return user, nil
	}

	user.UserName = newUsername
	if err := s.userRepository.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
