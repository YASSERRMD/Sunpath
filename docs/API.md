# Sunpath API

## GET /api/healthz

Health check endpoint.

**Response 200:**
```json
{"data": {"status": "ok"}}
```

## GET /api/horizon

Compute the horizon profile for a point. Returns 360 azimuth samples (one per degree, 0=N, 90=E, 180=S, 270=W).

**Query Parameters:**

| Param | Type   | Required | Default | Description                        |
|-------|--------|----------|---------|------------------------------------|
| lat   | float  | yes      | -       | Latitude (-90 to 90)               |
| lng   | float  | yes      | -       | Longitude (-180 to 180)            |
| h     | float  | no       | 1.5     | Observer height in metres          |

**Response 200:**
```json
{
  "data": {
    "horizon": [0.0, 0.0, ...],
    "lat": 48.8566,
    "lng": 2.3522,
    "observer_height": 1.5,
    "confidence": 0.85,
    "building_count": 42,
    "estimated_count": 6,
    "data_hash": "abc123..."
  }
}
```

**Error Responses:**

| Code | Description                    |
|------|--------------------------------|
| 400  | Missing or invalid parameters  |
| 405  | Method not allowed             |
| 502  | Failed to fetch building data  |
| 500  | Internal computation error     |

**Error shape:**
```json
{"error": "description"}
```

## GET /api/geocode

Proxy to Nominatim OpenStreetMap geocoding API.

**Query Parameters:**

| Param | Type   | Required | Description          |
|-------|--------|----------|----------------------|
| q     | string | yes      | Free-form query      |

**Response 200:**
```json
{
  "data": [
    {
      "display_name": "Paris, France",
      "lat": "48.8566",
      "lon": "2.3522",
      "type": "city",
      "importance": 0.95
    }
  ]
}
```

**Error Responses:**

| Code | Description                        |
|------|------------------------------------|
| 400  | Missing query parameter            |
| 405  | Method not allowed                 |
| 502  | Nominatim service unavailable      |
