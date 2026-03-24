package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var DefaultParams = &Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func HashPassword(password string, params *Params) (string, error) {
	if params == nil {
		params = DefaultParams
	}
	salt := make([]byte, params.SaltLength)

	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encoded := fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	)
	return encoded, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 5 {
		slog.Error("invalid hash format", "parts", len(parts))
		return false, fmt.Errorf("invalid hash format")
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8

	fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	newHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))
	if subtle.ConstantTimeCompare(hash, newHash) == 1 {
		return true, nil
	}
	return false, nil
}

func PasswordStrengthValidation(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	lower := regexp.MustCompile(`[a-z]`)
	upper := regexp.MustCompile(`[A-Z]`)
	number := regexp.MustCompile(`[0-9]`)
	special := regexp.MustCompile(`[@$!%*?&]`)

	if !lower.MatchString(password) {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !upper.MatchString(password) {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !number.MatchString(password) {
		return errors.New("password must contain at least one number")
	}
	if !special.MatchString(password) {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
