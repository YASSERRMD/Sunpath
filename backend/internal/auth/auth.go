package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func GenerateToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(raw)
	hash := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(hash[:]), nil
}

func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func GenerateMagicCode() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateMagicLinkEmail(from, to, code string) string {
	link := fmt.Sprintf("https://sunpath.app/auth/callback?code=%s", code)
	return fmt.Sprintf("From: %s\nTo: %s\nSubject: Sign in to Sunpath\n\nSign in to Sunpath by clicking this link:\n\n%s\n\nIf you did not request this, ignore this email.\n", from, to, link)
}
