package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

// opaque token
func GenerateToken() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(b), nil
}

func GenerateAuthCode() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GenerateSessionId() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateS256CodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func VerifyCodeChallenge(codeVerifier, codeChallenge, method string) error {
	switch method {
	case "S256":
		expectedChallenge := GenerateS256CodeChallenge(codeVerifier)
		if expectedChallenge != codeChallenge {
			return errors.New("invalid_code_verifier")
		}
	case "plain":
		if codeVerifier != codeChallenge {
			return errors.New("invalid_code_verifier")
		}
	default:
		return errors.New("unsupported_code_challenge_method")
	}
	return nil
}
