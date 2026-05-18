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

export async function fetchHorizon(lat: number, lng: number, h: number): Promise<HorizonProfile> {
  const params = new URLSearchParams({ lat: String(lat), lng: String(lng), h: String(h) })
  const res = await fetch(`/api/horizon?${params}`)
  const body: ApiResponse<HorizonProfile> = await res.json()
  if (body.error || !body.data) {
    throw new Error(body.error || 'failed to fetch horizon')
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
