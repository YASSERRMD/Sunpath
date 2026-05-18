# Sunpath

A solar exposure and shadow analysis application built on open-source maps.

Drop a pin on a map and discover how much direct sunlight that point receives across any day of the year.

## Stack

- **Backend**: Go (1.22+), SQLite (modernc.org/sqlite), Overpass API
- **Frontend**: React, Vite, TypeScript, MapLibre GL JS
- **Maps**: OpenStreetMap vector tiles (MapTiler or self-hosted)
- **Geocoding**: Nominatim (OpenStreetMap)

## Local development

Ensure you have Go 1.22+ and Node.js 18+ installed.

```bash
# 1. Start the backend
cd backend
go run ./cmd/sunpathd &

# 2. Start the frontend
cd frontend
cp ../.env.example .env
npm install
npm run dev
```

The app runs at http://localhost:5173. The backend serves at http://localhost:8080.

### Using docker-compose

```bash
docker-compose up -d
cp .env.example .env
cd frontend && npm install && npm run dev
```

### Configuration

Copy `.env.example` to `.env` and set:

| Variable | Default | Description |
|----------|---------|-------------|
| `TILE_STYLE_URL` | `https://demotiles.maplibre.org/style.json` | Map tile style URL |
| `OVERPASS_URL` | `https://overpass-api.de/api/interpreter` | Overpass API endpoint |
| `LISTEN_ADDR` | `:8080` | Backend listen address |
| `DB_PATH` | `sunpath.db` | SQLite database path |

## Project structure

```
sunpath/
  backend/        Go backend (horizon engine, API, store)
  frontend/       React + Vite frontend (map, heatmap, analysis)
  docs/           Architecture and API docs
```

## License

MIT
