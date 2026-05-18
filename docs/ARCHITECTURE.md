# Sunpath Architecture

## Shadow Model: 2.5D Core (DSM optional)

Buildings are treated as 2.5D extruded prisms from OSM footprints. Shadow casting is geometric ray intersection between the sun vector and these prisms.

DSM terrain/vegetation shadows are available as an optional overlay (opt-in). They do not change the core engine.

## Compute Split: Backend Precompute, Client Interpolation

**Backend (Go)**: Fetches and caches OSM building geometry, extrudes prisms, and computes a horizon profile — the obstruction angle of surrounding buildings for each compass azimuth (360 samples). This is the expensive step, cached per point in Postgres (with Redis hot cache).

**Client (browser)**: Given the cached horizon profile, the client computes sun position for any date/time (cheap astronomy math) and checks it against the horizon profile. Full year heatmap is thousands of cheap lookups, done instantly.

## The Horizon Profile (Core Abstraction)

```
horizon[az] = max elevation angle at which a building edge obstructs the sky in direction az
```

A point is in direct sun iff:
```
sun_elevation > horizon[round(sun_azimuth)] AND sun_elevation > 0
```
Azimuth convention: 0=N, 90=E, 180=S, 270=W (tested explicitly).

## Technology Stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Backend | Go (1.22+) | Scalable, fast |
| Geometry core | Go, custom ray casting | No heavyweight GIS dependency |
| Primary DB | Postgres 16 + PostGIS | Geospatial queries, system of record |
| Hot cache / queue | Redis 7 | Cache horizon profiles, job queue |
| OSM data | Overpass API | Open data, no key |
| Frontend | React + Vite + TypeScript + PWA | Reactive, offline-capable |
| Map | MapLibre GL JS | Open-source, no Mapbox token |
| Tiles | MapTiler / self-hosted | No proprietary token |
| Sun math | suncalc (client) + Go port | Well-tested astronomy |
| Charts | Hand-built canvas heatmap | Keeps bundle small |
| Geocoding | Nominatim (OSM) | Open, no key |
| Auth | Magic-link / single OAuth | No password storage |
| Migrations | goose | Up/down reversible |
| Containers | Docker Compose | Single command to run all services |

## Building Height Resolution from OSM

1. `height` tag (metres) — use directly.
2. `building:levels` tag — 3.2 m per level + 1 m base.
3. No data — configurable default (8 m), flagged as estimated.

Estimated fraction surfaced as confidence indicator.

## Repository Shape

```
sunpath/
  AGENTS.md
  README.md
  LICENSE
  .gitignore
  docker-compose.yml
  docker-compose.prod.yml
  backend/
    cmd/
      sunpathd/main.go     # server entrypoint
      paritycheck/          # cross-store parity check
    internal/
      geo/       # geo types, polygon ops, extrusion
      osm/       # Overpass client, building parsing
      horizon/   # horizon profile engine (unchanged)
      sun/       # solar position (validation copy)
      dsm/       # DSM terrain elevation overlay
      api/       # HTTP handlers, routing
      store/     # Storage interface + Postgres adapter
    migrations/  # goose SQL migrations
    go.mod
  frontend/
    src/
      lib/sun.ts         # client sun position + rule
      lib/horizon.ts     # horizon consumption, year compute
      components/        # Map, PinInspector, Heatmap, etc.
      workers/           # Web Worker for year computation
      App.tsx
    package.json
    vite.config.ts
  docs/
    ARCHITECTURE.md
    API.md
```

## Edge Cases

- Open field: all-zero horizon.
- Courtyard: heavy obstruction.
- High latitude: polar day / polar night.
- Antimeridian / timezone crossing: documented limitation.
- Observer above all buildings: near-zero horizon.
