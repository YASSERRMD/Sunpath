package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/dsm"
	"github.com/yasserrmd/sunpath/backend/internal/horizon"
	"github.com/yasserrmd/sunpath/backend/internal/osm"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

const evictionTTL = 7 * 24 * time.Hour

type Server struct {
	store         store.Storage
	overpassURL   string
	cachedClient  *osm.CachedClient
	horizonComp   *horizon.CachedComputer
	geoClient     *osm.RateLimitedClient
	elevClient    *dsm.ElevationClient
	errorCounts   map[string]*int64
}

func NewServer(st store.Storage, overpassURL string) *Server {
	oc := osm.NewClient(overpassURL)
	cc := osm.NewCachedClient(oc, st, osm.DefaultConfig())
	hc := horizon.NewCachedComputer(st)
	elevURL := os.Getenv("ELEVATION_API_URL")
	return &Server{
		store:        st,
		overpassURL:  overpassURL,
		cachedClient: cc,
		horizonComp:  hc,
		geoClient:    osm.NewRateLimitedClient(2),
		elevClient:   dsm.NewElevationClient(elevURL),
		errorCounts:  map[string]*int64{},
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", s.handleHealthz)
	mux.HandleFunc("/api/horizon", cors(s.handleHorizon))
	mux.HandleFunc("/api/buildings", cors(s.handleBuildings))
	mux.HandleFunc("/api/grid", cors(s.handleGrid))
	mux.HandleFunc("/api/metrics", cors(s.handleMetrics))
	mux.HandleFunc("/api/cache/evict", cors(s.handleCacheEvict))
	mux.HandleFunc("/api/geocode", cors(s.handleGeocode))
	return withLogging(mux)
}

type envelope struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) writeError(w http.ResponseWriter, status int, msg string) {
	key := http.StatusText(status)
	if key == "" {
		key = "unknown"
	}
	counter, ok := s.errorCounts[key]
	if !ok {
		var zero int64
		s.errorCounts[key] = &zero
		counter = &zero
	}
	atomic.AddInt64(counter, 1)
	writeJSON(w, status, envelope{Error: msg})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, envelope{Error: msg})
}

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "http://localhost:4173" || strings.HasSuffix(origin, ".sunpath.app") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}
		next(w, r)
	}
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
