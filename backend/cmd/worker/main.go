package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yasserrmd/sunpath/backend/internal/dsm"
	"github.com/yasserrmd/sunpath/backend/internal/geo"
	"github.com/yasserrmd/sunpath/backend/internal/horizon"
	"github.com/yasserrmd/sunpath/backend/internal/osm"
	"github.com/yasserrmd/sunpath/backend/internal/queue"
	"github.com/yasserrmd/sunpath/backend/internal/store"
)

func fetchBuildingsAround(point geo.Point, client *osm.CachedClient) ([]geo.Building, error) {
	lat0 := math.Floor(point.Lat*10) / 10
	lng0 := math.Floor(point.Lng*10) / 10
	return client.FetchBuildingsInBBox(lat0-0.05, lng0-0.05, lat0+0.05, lng0+0.05)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	databaseURL := getEnv("DATABASE_URL", "postgres://sunpath:sunpath@localhost:5432/sunpath?sslmode=disable")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	overpassURL := getEnv("OVERPASS_URL", "https://overpass-api.de/api/interpreter")

	st, err := store.NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("opening store: %v", err)
	}
	defer st.Close()

	q := queue.New(redisAddr, "", 0)
	defer q.Close()

	oc := osm.NewClient(overpassURL)
	cc := osm.NewCachedClient(oc, st, osm.DefaultConfig())
	hc := horizon.NewCachedComputer(st)

	elevURL := os.Getenv("ELEVATION_API_URL")
	var elevClient *dsm.ElevationClient
	if elevURL != "" {
		elevClient = dsm.NewElevationClient(elevURL)
	}

	log.Println("worker started, waiting for jobs...")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down worker...")
			return
		default:
		}

		job, err := q.Dequeue(context.Background(), 5*time.Second)
		if err != nil {
			log.Printf("dequeue error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if job == nil {
			continue
		}

		start := time.Now()
		result := &queue.JobResult{ID: job.ID}

		p := geo.Point{Lat: job.Lat, Lng: job.Lng}
		buildings, err := fetchBuildingsAround(p, cc)
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			log.Printf("job %s: fetch error: %v", job.ID, err)
		} else {
			var profile horizon.Profile
			if job.UseDSM && elevClient != nil {
				terrain, tErr := dsm.ComputeTerrainHorizon(elevClient, p.Lat, p.Lng, job.Height, 0)
				if tErr != nil {
					log.Printf("job %s: terrain compute error (falling back): %v", job.ID, tErr)
				} else {
					profile, err = hc.ComputeWithTerrain(p, job.Height, buildings, &terrain.Horizon)
				}
			}
			if !job.UseDSM || elevClient == nil {
				profile, err = hc.Compute(p, job.Height, buildings)
			}
			if err != nil {
				result.Status = "failed"
				result.Error = err.Error()
				log.Printf("job %s: compute error: %v", job.ID, err)
			} else {
				profileJSON, _ := json.Marshal(profile)
				result.Status = "completed"
				result.Profile = profileJSON
				result.Latency = float64(time.Since(start).Microseconds()) / 1000.0
				log.Printf("job %s: completed in %.1fms", job.ID, result.Latency)
			}
		}

		if err := q.StoreResult(context.Background(), result); err != nil {
			log.Printf("job %s: store result error: %v", job.ID, err)
		}
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
