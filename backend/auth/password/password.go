package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	z "github.com/Oudwins/zog"
	"golang.org/x/crypto/argon2"
)

// PasswordSchema defines the schema for a password.
var PasswordSchema = z.String().Trim().Min(6, z.Message("Password too short")).Max(20, z.Message("Password too long"))

// OnetimePasswordSchema defines the schema for a one-time password.
var OnetimePasswordSchema = z.String().Trim().Len(6, z.Message("Onetime password must be 6 digits")).Match(
	regexp.MustCompile(`^[0-9]{6}$`),
	z.Message("Onetime password must be 6 digits"),
)

// ErrInvalidPassword is returned when a password is invalid.
var ErrInvalidPassword = errors.New("invalid password")

var ErrPasswordHashing = errors.New("password hashing failed")

var ErrSaltGeneration = errors.New("salt generation failed")

type argon2Configuration struct {
	HashRaw    []byte
	Salt       []byte
	TimeCost   uint32
	MemoryCost uint32
	Threads    uint8
	KeyLength  uint32
}

func generateCryptographicSalt(saltSize uint32) ([]byte, error) {
	salt := make([]byte, saltSize)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func CreateArgon2idHash(password string) (string, error) {
	config := &argon2Configuration{
		TimeCost:   2,
		MemoryCost: 64 * 1024,
		Threads:    2,
		KeyLength:  32,
	}

	salt, err := generateCryptographicSalt(16)
	if err != nil {
		return "", ErrSaltGeneration
	}
	config.Salt = salt

	// Execute Argon2id hashing algorithm
	config.HashRaw = argon2.IDKey(
		[]byte(password),
		config.Salt,
		config.TimeCost,
		config.MemoryCost,
		config.Threads,
		config.KeyLength,
	)

	// Generate standardized hash format
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		config.MemoryCost,
		config.TimeCost,
		config.Threads,
		base64.RawStdEncoding.EncodeToString(config.Salt),
		base64.RawStdEncoding.EncodeToString(config.HashRaw),
	)

	return encodedHash, nil
}

func parseArgon2Hash(encodedHash string) (*argon2Configuration, error) {
	components := strings.Split(encodedHash, "$")
	if len(components) != 6 {
		return nil, errors.New("invalid hash format structure")
	}

	// Validate algorithm identifier
	if !strings.HasPrefix(components[1], "argon2id") {
		return nil, errors.New("unsupported algorithm variant")
	}

	// Extract version information
	var version int
	_, err := fmt.Sscanf(components[2], "v=%d", &version)
	if err != nil {
		return nil, fmt.Errorf("version parsing failed: %w", err)
	}

	// Parse configuration parameters
	config := &argon2Configuration{}
	_, err = fmt.Sscanf(components[3], "m=%d,t=%d,p=%d", &config.MemoryCost, &config.TimeCost, &config.Threads)
	if err != nil {
		return nil, fmt.Errorf("parameter parsing failed: %w", err)
	}

	// Decode salt component
	salt, err := base64.RawStdEncoding.DecodeString(components[4])
	if err != nil {
		return nil, fmt.Errorf("salt decoding failed: %w", err)
	}
	config.Salt = salt

	// Decode hash component
	hash, err := base64.RawStdEncoding.DecodeString(components[5])
	if err != nil {
		return nil, fmt.Errorf("hash decoding failed: %w", err)
	}
	config.HashRaw = hash
	config.KeyLength = uint32(len(hash))

	return config, nil
}

func VerifyPassword(correctPasswordHash, userProvidedPassword string) error {
	config, err := parseArgon2Hash(correctPasswordHash)
	if err != nil {
		return fmt.Errorf("hash parsing failed: %w", err)
	}

	// Generate hash using identical parameters
	computedHash := argon2.IDKey(
		[]byte(userProvidedPassword),
		config.Salt,
		config.TimeCost,
		config.MemoryCost,
		config.Threads,
		config.KeyLength,
	)

	// Perform constant-time comparison to prevent timing attacks
	match := subtle.ConstantTimeCompare(config.HashRaw, computedHash) == 1
	if !match {
		return ErrInvalidPassword
	}

	return nil
}

func GenerateOnetimePassword() (string, error) {
	const passwordLength = 6
	const charset = "0123456789"

	bytePassword := make([]byte, passwordLength)
	_, err := rand.Read(bytePassword)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes for onetime password: %w", err)
	}

	for i := 0; i < passwordLength; i++ {
		bytePassword[i] = charset[int(bytePassword[i])%len(charset)]
	}

	return string(bytePassword), nil
}
