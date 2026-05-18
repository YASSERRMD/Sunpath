package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

var ErrSessionNotFound = errors.New("session not found")
var ErrSessionExpired = errors.New("session expired")

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionStore interface {
	CreateUser(ctx context.Context, email, name string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id int64) (*User, error)
	CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*Session, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*Session, error)
	DeleteExpiredSessions(ctx context.Context) error
	CreateMagicLink(ctx context.Context, email, code string, expiresAt time.Time) error
	ConsumeMagicLink(ctx context.Context, code string) (*string, error)
}

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

func ValidateSession(ctx context.Context, store SessionStore, token string) (*User, error) {
	tokenHash := HashToken(token)
	session, err := store.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}
	user, err := store.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrSessionNotFound
	}
	return user, nil
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
