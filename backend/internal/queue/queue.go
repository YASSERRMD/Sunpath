package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const horizonQueueKey = "horizon:jobs"
const horizonResultPrefix = "horizon:result:"

type HorizonJob struct {
	ID     string  `json:"id"`
	Lat    float64 `json:"lat"`
	Lng    float64 `json:"lng"`
	Height float64 `json:"height"`
	UseDSM bool    `json:"use_dsm"`
}

type JobResult struct {
	ID      string  `json:"id"`
	Status  string  `json:"status"`
	Error   string  `json:"error,omitempty"`
	Profile []byte  `json:"profile,omitempty"`
	Latency float64 `json:"latency_ms,omitempty"`
}

type JobQueue struct {
	client *redis.Client
}

func New(addr, password string, db int) *JobQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &JobQueue{client: client}
}

func (q *JobQueue) Close() error {
	return q.client.Close()
}

func (q *JobQueue) Enqueue(ctx context.Context, job *HorizonJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return q.client.LPush(ctx, horizonQueueKey, data).Err()
}

func (q *JobQueue) Dequeue(ctx context.Context, timeout time.Duration) (*HorizonJob, error) {
	result, err := q.client.BRPop(ctx, timeout, horizonQueueKey).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(result) < 2 {
		return nil, fmt.Errorf("unexpected BRPop result: %v", result)
	}
	var job HorizonJob
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (q *JobQueue) StoreResult(ctx context.Context, result *JobResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	key := horizonResultPrefix + result.ID
	return q.client.Set(ctx, key, data, 24*time.Hour).Err()
}

func (q *JobQueue) GetResult(ctx context.Context, jobID string) (*JobResult, error) {
	key := horizonResultPrefix + jobID
	data, err := q.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var result JobResult
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (q *JobQueue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}
