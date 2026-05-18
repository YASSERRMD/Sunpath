package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/store"
)

type metricsResponse struct {
	Cache  store.CacheStats       `json:"cache"`
	Errors map[string]int64       `json:"errors"`
}

var (
	requestCount   atomic.Int64
	errorCount     atomic.Int64
	latencySumMs   atomic.Int64
	activeRequests atomic.Int64
)

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		writeJSON(w, 405, envelope{Error: "method not allowed"})
		return
	}

	cacheStats := s.store.Stats()
	errMap := make(map[string]int64, len(s.errorCounts))
	for code, counter := range s.errorCounts {
		errMap[code] = atomic.LoadInt64(counter)
	}

	count := requestCount.Load()
	var avg int64
	if count > 0 {
		avg = latencySumMs.Load() / count
	}

	writeJSON(w, 200, envelope{Data: map[string]interface{}{
		"http": map[string]int64{
			"request_count":   requestCount.Load(),
			"error_count":     errorCount.Load(),
			"active_requests": activeRequests.Load(),
			"avg_latency_ms":  avg,
		},
		"cache":        cacheStats,
		"error_counts": errMap,
	}})
}

type logEntry struct {
	Timestamp   string `json:"timestamp"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Status      int    `json:"status"`
	LatencyMs   int64  `json:"latency_ms"`
	UserAgent   string `json:"user_agent"`
	RemoteAddr  string `json:"remote_addr"`
}

func observedHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		activeRequests.Add(1)
		requestCount.Add(1)
		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, status: 200}
		next.ServeHTTP(sw, r)

		latency := time.Since(start)
		latencySumMs.Add(latency.Milliseconds())
		activeRequests.Add(-1)

		if sw.status >= 400 {
			errorCount.Add(1)
		}

		host := r.RemoteAddr
		if hn, err := os.Hostname(); err == nil {
			host = hn + " " + r.RemoteAddr
		}

		line, _ := json.Marshal(logEntry{
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     sw.status,
			LatencyMs:  latency.Milliseconds(),
			UserAgent:  r.UserAgent(),
			RemoteAddr: host,
		})
		log.Println(string(line))
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func (s *Server) handleCacheEvict(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		s.writeError(w, 405, "method not allowed")
		return
	}

	removed, err := s.store.EvictOlderThan(evictionTTL)
	if err != nil {
		s.writeError(w, 500, "eviction failed")
		return
	}
	writeJSON(w, 200, envelope{Data: map[string]int{"removed": removed}})
}
