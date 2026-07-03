package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
	"golang.org/x/crypto/argon2"
)

var PasswordSchema = z.String().Trim().Min(6, z.Message("Passwort zu kurz")).Max(72, z.Message("Passwort zu lang"))

var ErrPasswordTooWeak = errors.New("password too weak")

var ErrInvalidPassword = errors.New("invalid password")

var ErrNoPassword = errors.New("no password set")

// ErrOnetimePasswordLocked: das Einmalpasswort wurde nach zu vielen Fehlversuchen
// ungültig; der Admin muss ein neues erzeugen.
var ErrOnetimePasswordLocked = errors.New("onetime password locked after too many attempts")

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

func createArgon2idHash(password string) (string, error) {
	config := &argon2Configuration{
		TimeCost:   2,
		MemoryCost: 64 * 1024,
		Threads:    2,
		KeyLength:  32,
	}

	salt, err := generateCryptographicSalt(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate cryptographic salt: %w", err)
	}
	config.Salt = salt

	config.HashRaw = argon2.IDKey(
		[]byte(password),
		config.Salt,
		config.TimeCost,
		config.MemoryCost,
		config.Threads,
		config.KeyLength,
	)

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

	if !strings.HasPrefix(components[1], "argon2id") {
		return nil, errors.New("unsupported algorithm variant")
	}

	var version int
	_, err := fmt.Sscanf(components[2], "v=%d", &version)
	if err != nil {
		return nil, fmt.Errorf("version parsing failed: %w", err)
	}

	config := &argon2Configuration{}
	_, err = fmt.Sscanf(components[3], "m=%d,t=%d,p=%d", &config.MemoryCost, &config.TimeCost, &config.Threads)
	if err != nil {
		return nil, fmt.Errorf("parameter parsing failed: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(components[4])
	if err != nil {
		return nil, fmt.Errorf("salt decoding failed: %w", err)
	}
	config.Salt = salt

	hash, err := base64.RawStdEncoding.DecodeString(components[5])
	if err != nil {
		return nil, fmt.Errorf("hash decoding failed: %w", err)
	}
	config.HashRaw = hash
	config.KeyLength = uint32(len(hash))

	return config, nil
}

func verifyPassword(correctPasswordHash, userProvidedPassword string) error {
	config, err := parseArgon2Hash(correctPasswordHash)
	if err != nil {
		return fmt.Errorf("hash parsing failed: %w", err)
	}

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

func generateOnetimePassword() (string, error) {
	const passwordLength = 8
	// Exakt 32 eindeutige Zeichen (Kleinbuchstaben + Ziffern, ohne die verwechselbaren
	// o, i, l und 1; die 0 ist eindeutig, weil das o fehlt): 32^8 ≈ 1,1 · 10^12
	// Kombinationen statt 10^6 bei 6 Ziffern. Da 256 % 32 == 0, ist charset[b % 32]
	// exakt gleichverteilt (kein Modulo-Bias).
	const charset = "abcdefghjkmnpqrstuvwxyz023456789"

	bytePassword := make([]byte, passwordLength)
	_, err := rand.Read(bytePassword)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes for onetime password: %w", err)
	}

	for i := range passwordLength {
		bytePassword[i] = charset[int(bytePassword[i])%len(charset)]
	}

	return string(bytePassword), nil
}

func generateOnetimePasswordHash() (string, string, error) {
	onetimePassword, err := generateOnetimePassword()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate one-time password: %w", err)
	}

	onetimePasswordHash, err := createArgon2idHash(onetimePassword)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash one-time password: %w", err)
	}

	return onetimePassword, onetimePasswordHash, nil
}
