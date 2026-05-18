# Sunpath Architecture

## Shadow Model: 2.5D Core (DSM optional later)

Core (Phases 1-9): Buildings are treated as 2.5D extruded prisms. Each OSM building footprint polygon is extruded vertically to its height. Shadow casting is geometric ray intersection between the sun vector and these prisms.

A later, clearly isolated phase adds Digital Surface Model / terrain-and-vegetation shadows as an optional overlay. It must not be entangled with the core engine.

## Compute Split: Backend Precompute, Client Interpolation

**Backend (Go)**: Fetches and caches OSM building geometry, extrudes prisms, and for a requested point precomputes a horizon profile — the obstruction angle of surrounding buildings for each compass azimuth (360 samples, one per degree). This is the expensive step and it is cached per point.

**Client (browser)**: Given the cached horizon profile, the client computes sun position for any date/time (cheap astronomy math) and checks it against the horizon profile. Producing the full year heatmap is then thousands of cheap lookups, done instantly in the browser.

## The Horizon Profile (Core Abstraction)

For a point P at observer height h:

```
horizon[az] = maximum elevation angle at which a building edge obstructs the sky, looking in direction az
```

A point is in direct sun if and only if:

```
sun_elevation > horizon[round(sun_azimuth)] AND sun_elevation > 0
```

Azimuth convention: 0=N, 90=E, 180=S, 270=W (documented and tested explicitly).

## Technology Stack

| Layer | Choice | Reason |
|-------|--------|--------|
| Backend | Go (1.22+) | Scalable backend |
| Geometry core | Go, custom ray casting | No heavyweight GIS dependency |
| Backend cache | SQLite (modernc.org/sqlite, no CGO) | Single-file, zero operations |
| OSM data | Overpass API | Open data, no API key |
| Frontend | React + Vite + TypeScript | Reactive, low-latency frontend |
| Map rendering | MapLibre GL JS | Open-source, no Mapbox token |
| Tiles | MapTiler open tiles or self-hosted | No proprietary token |
| Sun math | suncalc (client) + Go port (backend) | Well-tested astronomy |
| Charts | Hand-built canvas heatmap | Keeps bundle small |
| Geocoding | Nominatim (OSM) | Open, no API key |

## Building Height Resolution from OSM

Priority order:
1. `height` tag (metres) — use directly.
2. `building:levels` tag — multiply by 3.2 m per level, plus 1 m base.
3. No data — apply configurable default (8 m for building) and flag as estimated.

The fraction of estimated-height buildings is surfaced as a confidence indicator.

## Repository Shape

```
sunpath/
  AGENTS.md
  README.md
  LICENSE
  .gitignore
  docker-compose.yml
  backend/
    cmd/sunpathd/main.go
    internal/
      geo/       # geo types, polygon ops, extrusion
      osm/       # Overpass client, building parsing, cache
      horizon/   # horizon profile engine
      sun/       # solar position (backend validation copy)
      api/       # HTTP handlers, routing
      store/     # SQLite cache layer
    go.mod
  frontend/
    src/
      lib/sun.ts         # client sun position + sun/shade rule
      lib/horizon.ts     # horizon profile consumption, year compute
      components/        # Map, PinInspector, YearHeatmap, etc.
      App.tsx
    package.json
    vite.config.ts
  docs/
    ARCHITECTURE.md
    API.md
```

## Edge Cases

- Open field (no buildings): horizon is all zeros.
- Courtyard surrounded by tall buildings: heavy obstruction.
- High latitude: polar day / polar night.
- Antimeridian / timezone crossing: documented limitation.
- Observer above all buildings: horizon collapses to near zero.
