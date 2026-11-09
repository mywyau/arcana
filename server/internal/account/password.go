package account

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory          = 19 * 1024
	argonIterations      = 2
	argonParallelism     = 1
	argonSaltBytes       = 16
	argonHashBytes       = 32
	minimumPasswordBytes = 8
	maximumPasswordBytes = 1024
)

var (
	ErrInvalidPassword     = errors.New("password is invalid")
	ErrInvalidPasswordHash = errors.New("invalid password hash")
)

type PasswordHasher struct{}

func NewPasswordHasher() PasswordHasher {
	return PasswordHasher{}
}

func (PasswordHasher) Hash(password string) (string, error) {
	if len(password) < minimumPasswordBytes {
		return "", fmt.Errorf("%w: must contain at least %d characters", ErrInvalidPassword, minimumPasswordBytes)
	}
	if len(password) > maximumPasswordBytes {
		return "", fmt.Errorf("%w: exceeds maximum length", ErrInvalidPassword)
	}
	salt := make([]byte, argonSaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	hash := argon2.IDKey(
		[]byte(password), salt,
		argonIterations, argonMemory, argonParallelism, argonHashBytes,
	)
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonIterations,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (PasswordHasher) Verify(encodedHash string, password string) (bool, error) {
	parameters, salt, want, err := parsePasswordHash(encodedHash)
	if err != nil {
		return false, err
	}
	if len(password) > maximumPasswordBytes {
		return false, nil
	}
	got := argon2.IDKey(
		[]byte(password), salt,
		parameters.iterations, parameters.memory, parameters.parallelism, uint32(len(want)),
	)
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

type argonParameters struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

func parsePasswordHash(encodedHash string) (argonParameters, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	if parts[2] != fmt.Sprintf("v=%d", version) {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	var parameters argonParameters
	if _, err := fmt.Sscanf(
		parts[3],
		"m=%d,t=%d,p=%d",
		&parameters.memory,
		&parameters.iterations,
		&parameters.parallelism,
	); err != nil || parameters.memory == 0 || parameters.memory > 1024*1024 ||
		parameters.iterations == 0 || parameters.iterations > 10 ||
		parameters.parallelism == 0 || parameters.parallelism > 16 {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	if parts[3] != fmt.Sprintf(
		"m=%d,t=%d,p=%d",
		parameters.memory,
		parameters.iterations,
		parameters.parallelism,
	) {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	hash, err := base64.RawStdEncoding.Strict().DecodeString(parts[5])
	if err != nil || len(hash) < 16 || len(hash) > 64 {
		return argonParameters{}, nil, nil, ErrInvalidPasswordHash
	}
	return parameters, salt, hash, nil
}
