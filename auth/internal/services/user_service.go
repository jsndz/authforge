package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
	"github.com/jsndz/authforge/pkg/email"
	"github.com/redis/go-redis/v9"
)

type UserService struct {
	userRepository *repository.UserRepository
	tokenService   *TokenService
	sessionService *SessionService
	emailService   *email.EmailService
	redis          *redis.Client
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
	SessionId    string       `json:"session_id"`
}

func NewUserService(
	repo *repository.UserRepository,
	tokenService *TokenService,
	sessionService *SessionService,
	emailService *email.EmailService,
	redis *redis.Client) *UserService {
	return &UserService{
		userRepository: repo,
		tokenService:   tokenService,
		sessionService: sessionService,
		emailService:   emailService,
		redis:          redis,
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

func (s *UserService) Login(ctx context.Context, useremail, password, ip string) (*LoginResponse, error) {
	if err := s.IsBlocked(ctx, useremail, ip); err != nil {
		return nil, err
	}
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
		_ = s.RecordLoginFailure(ctx, useremail, ip)
		return nil, errors.New("invalid credentials")
	}
	s.ResetLoginAttempts(ctx, useremail, ip)

	now := time.Now()

	user, err = s.userRepository.Update(user.ID, map[string]interface{}{
		"last_login_at": now,
	})
	if err != nil {
		return nil, err
	}

	accessToken, refreshToken, sessionId, err := s.sessionService.CreateSessionTokens(ctx, user.ID, "read write")
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
		SessionId:    sessionId,
	}, nil
}

func (s *UserService) Logout(ctx context.Context, RefreshToken string, accessToken string, sessionId string) error {
	s.sessionService.BlacklistToken(ctx, accessToken, 15*time.Minute)
	s.sessionService.RevokeSession(ctx, sessionId)
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

	_, err = s.userRepository.Update(user.ID, map[string]interface{}{
		"is_active":  false,
		"updated_at": time.Now(),
	})

	return err
}

func (s *UserService) VerifyEmail(ctx context.Context, rawToken string) (LoginResponse, error) {

	token, err := s.tokenService.VerifyToken(rawToken, model.TokenEmailVerification)
	if err != nil {
		log.Printf("Email verification failed: token verification error - %v", err)
		return LoginResponse{}, err
	}
	log.Printf("Token verified for user ID: %d", token.UserID)

	log.Printf("Updating user %d to mark email as verified", token.UserID)
	user, err := s.userRepository.Update(token.UserID, map[string]interface{}{
		"email_verified": true,
		"updated_at":     time.Now(),
	})

	if err != nil {
		log.Printf("Failed to update user %d: %v", token.UserID, err)
		return LoginResponse{}, err
	}
	log.Printf("User %d email verified successfully", token.UserID)

	log.Printf("Creating session tokens for user %d", token.UserID)
	accessToken, refreshToken, sessionId, err := s.sessionService.CreateSessionTokens(ctx, token.UserID, "read write")
	if err != nil {
		log.Printf("Failed to create session tokens for user %d: %v", token.UserID, err)
		return LoginResponse{}, err
	}
	log.Printf("Session tokens created successfully for user %d", token.UserID)

	return LoginResponse{
		User: UserResponse{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.UserName,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionId:    sessionId,
	}, nil
}

func (s *UserService) VerifyPasswordReset(ctx context.Context, rawToken string, password string) error {
	token, err := s.tokenService.VerifyToken(rawToken, model.TokenPasswordReset)
	if err != nil {
		return err
	}
	if err := security.PasswordStrengthValidation(password); err != nil {
		return err
	}
	passwordHash, err := security.HashPassword(password, security.DefaultParams)
	if err != nil {
		return err
	}
	_, err = s.userRepository.Update(token.UserID, map[string]interface{}{
		"password":   passwordHash,
		"updated_at": time.Now(),
	})
	if err != nil {
		return err
	}
	s.sessionService.AllSessionLogout(ctx, token.UserID)
	return nil
}

func (s *UserService) UpdateUsername(userID uint, username string) (*model.User, error) {
	newUsername := strings.TrimSpace(username)
	if newUsername == "" {
		return nil, errors.New("username is required")
	}

	user, err := s.userRepository.Update(userID, map[string]interface{}{
		"user_name":  newUsername,
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) IsBlocked(ctx context.Context, email, ip string) error {
	emailKey := "login:fail:user:" + email
	ipKey := "login:fail:ip:" + ip

	emailCount, err := s.redis.Get(ctx, emailKey).Int()
	if err == redis.Nil {
		emailCount = 0
	} else if err != nil {
		return err
	}

	ipCount, err := s.redis.Get(ctx, ipKey).Int()
	if err == redis.Nil {
		ipCount = 0
	} else if err != nil {
		return err
	}

	if emailCount >= 5 || ipCount >= 5 {
		return errors.New("too many login attempts, try again later")
	}

	return nil
}

func (s *UserService) RecordLoginFailure(ctx context.Context, email, ip string) error {
	emailKey := "login:fail:user:" + email
	ipKey := "login:fail:ip:" + ip

	count, err := s.redis.Incr(ctx, emailKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redis.Expire(ctx, emailKey, 10*time.Minute)
	}

	count, err = s.redis.Incr(ctx, ipKey).Result()
	if err != nil {
		return err
	}
	if count == 1 {
		s.redis.Expire(ctx, ipKey, 10*time.Minute)
	}

	return nil
}

func (s *UserService) ResetLoginAttempts(ctx context.Context, email, ip string) {
	s.redis.Del(ctx, "login:fail:user:"+email)
	s.redis.Del(ctx, "login:fail:ip:"+ip)
}

func (s *UserService) RequestPasswordReset(ctx context.Context, email string) error {
	user, err := s.userRepository.FindByEmail(email)
	if err != nil {
		return errors.New("if the email exists, a reset link will be sent")
	}

	token, err := s.tokenService.GetToken(user.ID, model.TokenPasswordReset)
	if err != nil {
		return err
	}

	return s.emailService.SendPasswordResetEmail(email, token)
}

func (s *UserService) CompleteLogout(ctx context.Context, userId uint) error {
	return s.sessionService.AllSessionLogout(ctx, userId)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (string, string, string, error) {
	userId, err := s.sessionService.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		s.sessionService.AllSessionLogout(ctx, userId)
		return "", "", "", errors.New("invalid refresh token")
	}

	err = s.sessionService.DeleteToken(ctx, refreshToken)
	if err != nil {
		return "", "", "", errors.New("failed to delete refresh token")
	}
	return s.sessionService.CreateSessionTokens(ctx, userId, "read write")
}

func (s *UserService) GetUserByID(userID uint) (*model.User, error) {
	return s.userRepository.FindByID(userID)
}

func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	return s.userRepository.FindByEmail(email)
}
func (s *UserService) CreateGoogleUser(email string) (*model.User, error) {
	user := &model.User{
		Email:         email,
		UserName:      email,
		IsActive:      true,
		EmailVerified: true,
	}

	err := s.userRepository.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}
