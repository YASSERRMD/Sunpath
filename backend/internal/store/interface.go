package store

import "time"

type Store interface {
	Close() error

	GetOSMExtract(bboxKey string) ([]BuildingRecord, error)
	PutOSMExtract(bboxKey string, buildings []BuildingRecord) error

	GetHorizonProfile(cacheKey string) (*HorizonRecord, error)
	PutHorizonProfile(cacheKey string, profile *HorizonRecord) error

	EvictOlderThan(age time.Duration) (int, error)
	Stats() CacheStats
}
