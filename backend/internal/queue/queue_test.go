package queue

import (
	"context"
	"testing"
	"time"
)

func skipIfNoRedis(t *testing.T) {
	t.Helper()
	q := New("localhost:6379", "", 0)
	if err := q.Ping(context.Background()); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	q.Close()
}

func TestEnqueueDequeue(t *testing.T) {
	skipIfNoRedis(t)

	q := New("localhost:6379", "", 0)
	defer q.Close()

	job := &HorizonJob{
		ID:     "test-1",
		Lat:    48.8566,
		Lng:    2.3522,
		Height: 1.5,
		UseDSM: false,
	}

	ctx := context.Background()
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	got, err := q.Dequeue(ctx, time.Second)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if got == nil {
		t.Fatal("expected job, got nil")
	}
	if got.ID != "test-1" || got.Lat != 48.8566 || got.Lng != 2.3522 {
		t.Errorf("job mismatch: %+v", got)
	}
	if got.Height != 1.5 {
		t.Errorf("height mismatch: %v", got.Height)
	}
}

func TestEnqueueDequeueWithDSM(t *testing.T) {
	skipIfNoRedis(t)

	q := New("localhost:6379", "", 0)
	defer q.Close()

	job := &HorizonJob{
		ID:     "test-dsm-1",
		Lat:    40.7484,
		Lng:    -73.9857,
		Height: 10.0,
		UseDSM: true,
	}

	ctx := context.Background()
	if err := q.Enqueue(ctx, job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	got, err := q.Dequeue(ctx, time.Second)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if got == nil {
		t.Fatal("expected job")
	}
	if !got.UseDSM {
		t.Error("expected UseDSM true")
	}
}

func TestStoreAndGetResult(t *testing.T) {
	skipIfNoRedis(t)

	q := New("localhost:6379", "", 0)
	defer q.Close()

	ctx := context.Background()
	result := &JobResult{
		ID:      "result-1",
		Status:  "completed",
		Profile: []byte(`{"lat":48.85}`),
		Latency: 123.4,
	}

	if err := q.StoreResult(ctx, result); err != nil {
		t.Fatalf("store result: %v", err)
	}

	got, err := q.GetResult(ctx, "result-1")
	if err != nil {
		t.Fatalf("get result: %v", err)
	}
	if got == nil {
		t.Fatal("expected result")
	}
	if got.Status != "completed" || got.Latency != 123.4 {
		t.Errorf("result mismatch: %+v", got)
	}
}

func TestGetNonExistentResult(t *testing.T) {
	skipIfNoRedis(t)

	q := New("localhost:6379", "", 0)
	defer q.Close()

	got, err := q.GetResult(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("get result: %v", err)
	}
	if got != nil {
		t.Error("expected nil for non-existent result")
	}
}

func TestDequeueTimeout(t *testing.T) {
	skipIfNoRedis(t)

	q := New("localhost:6379", "", 0)
	defer q.Close()

	got, err := q.Dequeue(context.Background(), time.Millisecond*100)
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if got != nil {
		t.Error("expected nil on timeout")
	}
}
