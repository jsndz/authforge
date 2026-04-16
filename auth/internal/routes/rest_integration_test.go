package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/jsndz/authforge/internal/bootstrap"
	"github.com/jsndz/authforge/internal/model"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/routes"
	"github.com/jsndz/authforge/internal/security"
	"github.com/jsndz/authforge/internal/services"
	"github.com/jsndz/authforge/pkg/db"
	"github.com/jsndz/authforge/pkg/util"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const testJWTSecret = "integration-test-secret"

type testApp struct {
	router         *gin.Engine
	gormDB         *gorm.DB
	redisClient    *redis.Client
	miniRedis      *miniredis.Miniredis
	sessionService *services.SessionService
	tokenService   *services.TokenService
}

func setupTestApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	gormDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.MigrateDB(gormDB); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})

	app := bootstrap.InitApp(gormDB, redisClient, testJWTSecret)
	router := gin.New()
	api := router.Group("/api/v1/auth")
	routes.AuthRouter(api, app.UserHandler, app.TokenHandler, testJWTSecret)

	tokenRepo := repository.NewTokenRepository(gormDB)
	tokenService := services.NewTokenService(tokenRepo)
	sessionService := services.NewSessionService(testJWTSecret, redisClient)

	t.Cleanup(func() {
		_ = redisClient.Close()
		mini.Close()
	})

	return &testApp{
		router:         router,
		gormDB:         gormDB,
		redisClient:    redisClient,
		miniRedis:      mini,
		sessionService: sessionService,
		tokenService:   tokenService,
	}
}

func performJSONRequest(t *testing.T, router *gin.Engine, method, path string, payload any, headers map[string]string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	body := bytes.NewBuffer(nil)
	if payload != nil {
		if err := json.NewEncoder(body).Encode(payload); err != nil {
			t.Fatalf("failed to encode payload: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func createUser(t *testing.T, gormDB *gorm.DB, username, email, password string, emailVerified bool) *model.User {
	t.Helper()

	hash, err := security.HashPassword(password, security.DefaultParams)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := &model.User{
		UserName:      username,
		Email:         email,
		Password:      hash,
		IsActive:      true,
		EmailVerified: emailVerified,
	}

	if err := gormDB.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return user
}

func TestAuthSignupEndpoint(t *testing.T) {
	tApp := setupTestApp(t)

	payload := map[string]string{
		"username": "signup_user",
		"email":    "signup@example.com",
		"password": "Str0ng@Pass",
	}

	resp := performJSONRequest(t, tApp.router, http.MethodPost, "/api/v1/auth/signup", payload, nil)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
	}

	var user model.User
	if err := tApp.gormDB.Where("email = ?", "signup@example.com").First(&user).Error; err != nil {
		t.Fatalf("expected user to be persisted: %v", err)
	}
}

func TestAuthLoginEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	createUser(t, tApp.gormDB, "login_user", "login@example.com", "Str0ng@Pass", true)

	payload := map[string]string{
		"email":    "login@example.com",
		"password": "Str0ng@Pass",
	}

	resp := performJSONRequest(t, tApp.router, http.MethodPost, "/api/v1/auth/login", payload, nil)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["access_token"] == "" {
		t.Fatalf("expected access_token in response")
	}

	foundRefreshCookie := false
	for _, cookie := range resp.Result().Cookies() {
		if cookie.Name == "refresh_token" {
			foundRefreshCookie = true
			break
		}
	}
	if !foundRefreshCookie {
		t.Fatalf("expected refresh_token cookie")
	}
}

func TestAuthVerifyEmailEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	user := createUser(t, tApp.gormDB, "verify_user", "verify@example.com", "Str0ng@Pass", false)

	token, err := tApp.tokenService.GetToken(user.ID, model.TokenEmailVerification)
	if err != nil {
		t.Fatalf("failed to generate email verification token: %v", err)
	}

	path := fmt.Sprintf("/api/v1/auth/email/verify?token=%s", token)
	resp := performJSONRequest(t, tApp.router, http.MethodGet, path, nil, nil)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected %d got %d: %s", http.StatusCreated, resp.Code, resp.Body.String())
	}

	var updated model.User
	if err := tApp.gormDB.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("failed to load updated user: %v", err)
	}
	if !updated.EmailVerified {
		t.Fatalf("expected user email to be verified")
	}
}

func TestAuthUpdateUsernameEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	user := createUser(t, tApp.gormDB, "update_user", "update@example.com", "Str0ng@Pass", true)

	accessToken, err := util.CreateJWT(user.ID, 15*time.Minute, testJWTSecret)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	payload := map[string]string{"username": "updated_name"}
	headers := map[string]string{"Authorization": "Bearer " + accessToken}

	resp := performJSONRequest(t, tApp.router, http.MethodPatch, "/api/v1/auth/update/username", payload, headers)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var updated model.User
	if err := tApp.gormDB.First(&updated, user.ID).Error; err != nil {
		t.Fatalf("failed to load updated user: %v", err)
	}
	if updated.UserName != "updated_name" {
		t.Fatalf("expected username to be updated")
	}
}

func TestAuthRequestPasswordResetEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	createUser(t, tApp.gormDB, "reset_user", "reset@example.com", "Str0ng@Pass", true)

	payload := map[string]string{"email": "reset@example.com"}
	resp := performJSONRequest(t, tApp.router, http.MethodPost, "/api/v1/auth/reset/password", payload, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAuthLogoutEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	user := createUser(t, tApp.gormDB, "logout_user", "logout@example.com", "Str0ng@Pass", true)

	ctx := context.Background()
	accessToken, refreshToken, err := tApp.sessionService.CreateSessionTokens(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to create session tokens: %v", err)
	}

	headers := map[string]string{"Authorization": "Bearer " + accessToken}
	cookie := &http.Cookie{Name: "refresh_token", Value: refreshToken}

	resp := performJSONRequest(t, tApp.router, http.MethodGet, "/api/v1/auth/logout", nil, headers, cookie)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAuthCompleteLogoutEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	user := createUser(t, tApp.gormDB, "logout_all_user", "logoutall@example.com", "Str0ng@Pass", true)

	accessToken, err := util.CreateJWT(user.ID, 15*time.Minute, testJWTSecret)
	if err != nil {
		t.Fatalf("failed to create access token: %v", err)
	}

	headers := map[string]string{"Authorization": "Bearer " + accessToken}
	resp := performJSONRequest(t, tApp.router, http.MethodGet, "/api/v1/auth/logout/all", nil, headers)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}
}

func TestAuthRefreshEndpoint(t *testing.T) {
	tApp := setupTestApp(t)
	user := createUser(t, tApp.gormDB, "refresh_user", "refresh@example.com", "Str0ng@Pass", true)

	ctx := context.Background()
	accessToken, refreshToken, err := tApp.sessionService.CreateSessionTokens(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to create session tokens: %v", err)
	}

	headers := map[string]string{"Authorization": "Bearer " + accessToken}
	cookie := &http.Cookie{Name: "refresh_token", Value: refreshToken}

	resp := performJSONRequest(t, tApp.router, http.MethodPost, "/api/v1/auth/refresh", nil, headers, cookie)
	if resp.Code != http.StatusOK {
		t.Fatalf("expected %d got %d: %s", http.StatusOK, resp.Code, resp.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["access_token"] == "" {
		t.Fatalf("expected access_token in response")
	}
}
