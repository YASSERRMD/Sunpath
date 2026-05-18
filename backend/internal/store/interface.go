package store

import (
	"context"
	"time"
)

type Storage interface {
	Close() error

	GetOSMExtract(bboxKey string) ([]BuildingRecord, error)
	PutOSMExtract(bboxKey string, buildings []BuildingRecord) error

	GetHorizonProfile(cacheKey string) (*HorizonRecord, error)
	PutHorizonProfile(cacheKey string, profile *HorizonRecord) error

	EvictOlderThan(age time.Duration) (int, error)
	Stats() CacheStats

	CreateUser(ctx context.Context, email, name string) (*UserRecord, error)
	GetUserByEmail(ctx context.Context, email string) (*UserRecord, error)
	GetUserByID(ctx context.Context, id int64) (*UserRecord, error)
	CreateSession(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time) (*SessionRecord, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*SessionRecord, error)
	DeleteExpiredSessions(ctx context.Context) error
	CreateMagicLink(ctx context.Context, email, code string, expiresAt time.Time) error
	ConsumeMagicLink(ctx context.Context, code string) (*string, error)
}
