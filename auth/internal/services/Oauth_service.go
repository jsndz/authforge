package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jsndz/authforge/internal/repository"
	"github.com/jsndz/authforge/internal/security"
	"github.com/jsndz/authforge/pkg/util"
	"github.com/redis/go-redis/v9"
)

type OauthService struct {
	oauthRepo            *repository.OauthRepo
	sessionService       *SessionService
	UserService          *UserService
	redis                *redis.Client
	jwtSecret            string
	GOOGLE_CLIENT_ID     string
	GOOGLE_CLIENT_SECRET string
	GOOGLE_CALLBACK_URL  string
}

type OAuthUser struct {
	UserID              uint
	ClientId            string
	Scope               string
	CodeChallenge       string
	CodeChallengeMethod string
	RedirectUri         string
	Email               string
}

func NewOAuthService(oauthRepo *repository.OauthRepo, sessionService *SessionService, userService *UserService, redis *redis.Client, jwtSecret string, GOOGLE_CLIENT_ID string, GOOGLE_CLIENT_SECRET string, GOOGLE_CALLBACK_URL string) *OauthService {
	return &OauthService{
		redis:                redis,
		oauthRepo:            oauthRepo,
		sessionService:       sessionService,
		UserService:          userService,
		jwtSecret:            jwtSecret,
		GOOGLE_CLIENT_ID:     GOOGLE_CLIENT_ID,
		GOOGLE_CLIENT_SECRET: GOOGLE_CLIENT_SECRET,
		GOOGLE_CALLBACK_URL:  GOOGLE_CALLBACK_URL,
	}
}

func (s *OauthService) AuthorizeClient(ctx context.Context, clientId, sessionId, redirectUri, scope, codeChallenge, codeChallengeMethod string) (string, error) {
	userId, err := s.sessionService.ValidateSession(ctx, sessionId)
	if err != nil {
		return "", err
	}

	client, err := s.oauthRepo.Get(clientId)
	if err != nil {
		if err.Error() == "client not found" {
			return "", fmt.Errorf("invalid client_id")
		}
		return "", err
	}
	if client.RedirectUri != redirectUri {
		return "", fmt.Errorf("invalid redirect URI")
	}
	allowedScope := strings.Split(client.Scopes, " ")
	requestedScope := strings.Split(scope, " ")
	if !isSubset(requestedScope, allowedScope) {
		return "", fmt.Errorf("scope not permitted")
	}
	authCode, err := security.GenerateAuthCode()
	if err != nil {
		return "", err
	}
	user, err := s.UserService.GetUserByID(userId)
	if err != nil {
		return "", err
	}
	oauthUser := &OAuthUser{
		UserID:              userId,
		ClientId:            clientId,
		Scope:               scope,
		CodeChallenge:       codeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		RedirectUri:         redirectUri,
		Email:               user.Email,
	}
	data, err := json.Marshal(oauthUser)
	if err != nil {
		return "", err
	}

	err = s.redis.Set(ctx, "auth_code:"+authCode, data, time.Minute*5).Err()
	if err != nil {
		return "", err
	}
	return authCode, nil
}

func (s *OauthService) Token(ctx context.Context, clientId, code, redirectUri, codeVerifier string) (string, string, string, string, error) {
	var user OAuthUser
	val, err := s.redis.Get(ctx, "auth_code:"+code).Result()
	if err == redis.Nil {
		return "", "", "", "", errors.New("invalid_or_expired_code")
	}
	if err != nil {
		return "", "", "", "", err
	}
	err = json.Unmarshal([]byte(val), &user)
	if err != nil {
		return "", "", "", "", err
	}
	if user.ClientId != clientId {
		return "", "", "", "", errors.New("invalid_client")
	}
	if user.RedirectUri != redirectUri {
		return "", "", "", "", errors.New("invalid_redirect_uri")
	}
	if user.CodeChallenge == "" || codeVerifier == "" {
		return "", "", "", "", errors.New("pkce_required")
	}
	err = security.VerifyCodeChallenge(codeVerifier, user.CodeChallenge, user.CodeChallengeMethod)
	if err != nil {
		return "", "", "", "", errors.New("invalid_code_verifier")
	}
	if err := s.redis.Del(ctx, code).Err(); err != nil {
		return "", "", "", "", err
	}

	access_token, refresh_token, session_id, err := s.sessionService.CreateSessionTokens(ctx, user.UserID, user.Scope)
	IDtoken := ""
	if strings.Contains(user.Scope, "openid") {
		IDtoken, err = util.CreateIDToken(user.UserID, user.Email, user.ClientId, s.jwtSecret)
		if err != nil {
			return "", "", "", "", err
		}
	}
	if err != nil {
		return "", "", "", "", err
	}

	return access_token, refresh_token, session_id, IDtoken, nil
}

func isSubset(requested, allowed []string) bool {
	m := make(map[string]bool)
	for _, s := range allowed {
		m[s] = true
	}
	for _, s := range requested {
		if !m[s] {
			return false
		}
	}
	return true
}

func (s *OauthService) Oauthlogin(redirectUri, clientId, state string) (string, error) {
	oauthClient, err := s.oauthRepo.Get(clientId)
	if err != nil {
		return "", err
	}
	if redirectUri != oauthClient.RedirectUri {
		return "", errors.New("invalid_redirect_uri")
	}
	encodedState := clientId + "|" + redirectUri + "|" + state
	url := "https://accounts.google.com/o/oauth2/v2/auth?" +
		"client_id=" + s.GOOGLE_CLIENT_ID +
		"&redirect_uri=" + s.GOOGLE_CALLBACK_URL +
		"&response_type=code" +
		"&scope=openid email profile" +
		"&access_type=offline" +
		"&prompt=consent" +
		"&state=" + encodedState

	return url, nil
}
func (s *OauthService) HandleGoogleCallback(ctx context.Context, code, state string) (string, string, string, string, error) {

	// 1. Exchange code → Google tokens
	resp, err := http.Post(
		"https://oauth2.googleapis.com/token",
		"application/x-www-form-urlencoded",
		strings.NewReader(
			"code="+code+
				"&client_id="+s.GOOGLE_CLIENT_ID+
				"&client_secret="+s.GOOGLE_CLIENT_SECRET+
				"&redirect_uri="+s.GOOGLE_CALLBACK_URL+
				"&grant_type=authorization_code",
		),
	)
	if err != nil {
		return "", "", "", "", err
	}
	defer resp.Body.Close()

	var googleResp struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return "", "", "", "", err
	}

	// 2. Decode id_token
	claims := jwt.MapClaims{}
	_, _, err = new(jwt.Parser).ParseUnverified(googleResp.IDToken, claims)
	if err != nil {
		return "", "", "", "", err
	}

	email := claims["email"].(string)

	// 3. Create / find user
	user, err := s.UserService.GetUserByEmail(email)
	if err != nil {
		user, err = s.UserService.CreateGoogleUser(email)
		if err != nil {
			return "", "", "", "", err
		}
	}

	access, refresh, session, err := s.sessionService.CreateSessionTokens(ctx, user.ID, "openid email profile")
	if err != nil {
		return "", "", "", "", err
	}
	parts := strings.Split(state, "|")
	if len(parts) != 3 {
		return "", "", "", "", errors.New("invalid_state")
	}

	clientId := parts[0]
	redirectUri := parts[1]
	originalState := parts[2]

	authCode, err := s.AuthorizeClient(ctx, clientId, session, redirectUri, "openid email profile", "", "")
	if err != nil {
		return "", "", "", "", err
	}

	u, _ := url.Parse(redirectUri)
	q := u.Query()
	q.Set("code", authCode)
	q.Set("state", originalState)
	u.RawQuery = q.Encode()

	return u.String(), access, refresh, session, nil
}
