package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	var version int
	var memory, iterations uint32
	var threads uint8
	var saltB64, hashB64 string

	_, err := fmt.Sscanf(encoded,
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&version, &memory, &iterations, &threads, &saltB64, &hashB64,
	)
	if err != nil {
		return false, fmt.Errorf("parse hash: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	computed := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(storedHash)))

	// constant-time compare
	if len(computed) != len(storedHash) {
		return false, nil
	}
	var diff byte
	for i := range computed {
		diff |= computed[i] ^ storedHash[i]
	}
	return diff == 0, nil
}
