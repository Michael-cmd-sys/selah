package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

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
	// Format: $argon2id$v=<v>$m=<m>,t=<t>,p=<p>$<salt>$<hash>
	// Split on "$" gives: ["", "argon2id", "v=...", "m=...,t=...,p=...", "<salt>", "<hash>"]
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, fmt.Errorf("parse hash: invalid format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parse hash version: %w", err)
	}

	var memory, iterations uint32
	var threads uint8
	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		return false, fmt.Errorf("parse hash: invalid params")
	}
	m, err := strconv.ParseUint(strings.TrimPrefix(paramParts[0], "m="), 10, 32)
	if err != nil {
		return false, fmt.Errorf("parse hash memory: %w", err)
	}
	ti, err := strconv.ParseUint(strings.TrimPrefix(paramParts[1], "t="), 10, 32)
	if err != nil {
		return false, fmt.Errorf("parse hash time: %w", err)
	}
	p, err := strconv.ParseUint(strings.TrimPrefix(paramParts[2], "p="), 10, 8)
	if err != nil {
		return false, fmt.Errorf("parse hash threads: %w", err)
	}
	memory, iterations, threads = uint32(m), uint32(ti), uint8(p)

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decode salt: %w", err)
	}
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decode hash: %w", err)
	}

	computed := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(storedHash)))

	if len(computed) != len(storedHash) {
		return false, nil
	}
	var diff byte
	for i := range computed {
		diff |= computed[i] ^ storedHash[i]
	}
	return diff == 0, nil
}
