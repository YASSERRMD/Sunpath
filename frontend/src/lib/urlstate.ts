export interface URLState {
  lat: number
  lng: number
  h: number
}

export function encodeState(lat: number, lng: number, h: number): string {
  const params = new URLSearchParams({
    lat: lat.toFixed(6),
    lng: lng.toFixed(6),
    h: h.toFixed(1),
  })
  return `?${params.toString()}`
}

export function decodeState(): URLState | null {
  const params = new URLSearchParams(window.location.search)
  const latStr = params.get('lat')
  const lngStr = params.get('lng')
  const hStr = params.get('h')

  if (!latStr || !lngStr) return null

  const lat = parseFloat(latStr)
  const lng = parseFloat(lngStr)
  const h = hStr ? parseFloat(hStr) : 1.5

  if (isNaN(lat) || isNaN(lng) || Math.abs(lat) > 90 || Math.abs(lng) > 180) {
    return null
  }

  return { lat, lng, h }
}

export function updateURL(lat: number, lng: number, h: number): void {
  const state = encodeState(lat, lng, h)
  window.history.replaceState(null, '', state)
}
