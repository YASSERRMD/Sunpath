export interface HorizonProfile {
  horizon: number[]
  lat: number
  lng: number
  observer_height: number
  confidence: number
  building_count: number
  estimated_count: number
  data_hash: string
}

export interface GeocodeResult {
  display_name: string
  lat: string
  lon: string
  type: string
  importance: number
}

interface ApiResponse<T> {
  data?: T
  error?: string
}

export async function fetchHorizon(lat: number, lng: number, h: number, useDSM?: boolean): Promise<HorizonProfile> {
  const params = new URLSearchParams({ lat: String(lat), lng: String(lng), h: String(h) })
  if (useDSM) params.set('dsm', 'true')
  const res = await fetch(`/api/horizon?${params}`)
  const body: ApiResponse<HorizonProfile> = await res.json()
  if (body.error || !body.data) {
    throw new Error(body.error || 'failed to fetch horizon')
  }
  return body.data
}

export interface BuildingOutline {
  osm_id: number
  height: number
  shape: [number, number][]
}

export async function fetchBuildings(lat: number, lng: number): Promise<BuildingOutline[]> {
  const params = new URLSearchParams({ lat: String(lat), lng: String(lng) })
  const res = await fetch(`/api/buildings?${params}`)
  const body: ApiResponse<BuildingOutline[]> = await res.json()
  if (body.error || !body.data) {
    throw new Error(body.error || 'failed to fetch buildings')
  }
  return body.data
}

export async function fetchGeocode(query: string): Promise<GeocodeResult[]> {
  const res = await fetch(`/api/geocode?q=${encodeURIComponent(query)}`)
  const body: ApiResponse<GeocodeResult[]> = await res.json()
  if (body.error || !body.data) {
    throw new Error(body.error || 'failed to geocode')
  }
  return body.data
}

export interface GridCell {
  lat: number
  lng: number
  sun_minutes: number
}

export async function fetchGrid(lat1: number, lng1: number, lat2: number, lng2: number, h: number, res?: number): Promise<GridCell[]> {
  const params = new URLSearchParams({
    lat1: String(lat1), lng1: String(lng1),
    lat2: String(lat2), lng2: String(lng2),
    h: String(h),
  })
  if (res) params.set('res', String(res))
  const result = await fetch(`/api/grid?${params}`)
  const body: ApiResponse<GridCell[]> = await result.json()
  if (body.error || !body.data) {
    throw new Error(body.error || 'failed to fetch grid')
  }
  return body.data
}
