package novelai

import (
	"encoding/base64"
	"encoding/hex"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/blake2b"
)

func DeriveKeys(email, password string) (string, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	accessSalt, err := deriveSalt(password, normalized, "novelai_data_access_key")
	if err != nil {
		return "", "", err
	}
	encryptionSalt, err := deriveSalt(password, normalized, "novelai_data_encryption_key")
	if err != nil {
		return "", "", err
	}

	passwordBytes := []byte(password)
	// Match the frontend structure: Argon2id with opslimit=2 and memlimit≈2e6 bytes.
	accessRaw := argon2.IDKey(passwordBytes, accessSalt, 2, 1953, 1, 64)
	encryptionRaw := argon2.IDKey(passwordBytes, encryptionSalt, 2, 1953, 1, 128)

	accessKey := base64.StdEncoding.EncodeToString(accessRaw)
	if len(accessKey) > 64 {
		accessKey = accessKey[:64]
	}
	encryptionKey := base64.StdEncoding.EncodeToString(encryptionRaw)
	return accessKey, encryptionKey, nil
}

func deriveSalt(password, normalizedEmail, suffix string) ([]byte, error) {
	prefix := password
	if len(prefix) > 6 {
		prefix = prefix[:6]
	}
	return keyedHash16(prefix + normalizedEmail + suffix)
}

func keyedHash16(input string) ([]byte, error) {
	sum, err := blake2b.New(16, nil)
	if err != nil {
		return nil, err
	}
	if _, err := sum.Write([]byte(input)); err != nil {
		return nil, err
	}
	return sum.Sum(nil), nil
}

func DeriveKeyDebug(email, password string) (string, string, string, string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	accessSalt, err := deriveSalt(password, normalized, "novelai_data_access_key")
	if err != nil {
		return "", "", "", "", err
	}
	encryptionSalt, err := deriveSalt(password, normalized, "novelai_data_encryption_key")
	if err != nil {
		return "", "", "", "", err
	}
	accessKey, encryptionKey, err := DeriveKeys(email, password)
	if err != nil {
		return "", "", "", "", err
	}
	return accessKey, encryptionKey, hex.EncodeToString(accessSalt), hex.EncodeToString(encryptionSalt), nil
}
