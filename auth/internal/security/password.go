package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log/slog"
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
	slog.Debug("hashing password")
	if params == nil {
		params = DefaultParams
	}
	salt := make([]byte, params.SaltLength)

	_, err := rand.Read(salt)
	if err != nil {
		slog.Error("failed to generate salt", "error", err)
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
	slog.Debug("password hashed successfully", "encoded", encoded)
	return encoded, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	slog.Debug("verifying password")
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
		slog.Error("failed to decode salt", "error", err)
		return false, err
	}

	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		slog.Error("failed to decode hash", "error", err)
		return false, err
	}
	newHash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(hash)))
	if subtle.ConstantTimeCompare(hash, newHash) == 1 {
		slog.Debug("password verification succeeded")
		return true, nil
	}
	slog.Debug("password verification failed")
	return false, nil
}
