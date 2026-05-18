import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'
import type { BuildingOutline } from '../lib/api'
import { getSunPosition } from '../lib/sun'

interface ShadowOverlayProps {
  map: maplibregl.Map | null
  date: Date
  lat: number
  lng: number
  buildings: BuildingOutline[]
}

export default function ShadowOverlay({ map, date, lat, lng, buildings }: ShadowOverlayProps) {
  const added = useRef(false)

  useEffect(() => {
    if (!map || buildings.length === 0) return

    const { azimuth, elevation } = getSunPosition(date, lat, lng)

    if (elevation <= 0) {
      if (added.current && map.getLayer('shadows')) {
        map.setLayoutProperty('shadows', 'visibility', 'none')
      }
      return
    }

    const azRad = azimuth * Math.PI / 180
    const elRad = elevation * Math.PI / 180
    const shadowLen = 200
    const dx = -Math.sin(azRad) * shadowLen / Math.tan(elRad)
    const dy = Math.cos(azRad) * shadowLen / Math.tan(elRad)

    const features: GeoJSON.Feature[] = []

    for (const b of buildings) {
      const projected = b.shape.map(([lng, lat]) => {
        const coords = projectPoint(lng, lat, lng, lat)
        const sx = coords[0] + dx
        const sy = coords[1] + dy
        const [plng, plat] = unprojectPoint(sx, sy, lng, lat)
        return [plng, plat] as [number, number]
      })

      features.push({
        type: 'Feature',
        properties: {},
        geometry: {
          type: 'Polygon',
          coordinates: [projected],
        },
      })
    }

    const source = map.getSource('shadows') as maplibregl.GeoJSONSource
    if (source) {
      source.setData({ type: 'FeatureCollection', features })
    } else {
      map.addSource('shadows', {
        type: 'geojson',
        data: { type: 'FeatureCollection', features },
      })
      map.addLayer({
        id: 'shadows',
        type: 'fill',
        source: 'shadows',
        paint: {
          'fill-color': 'rgba(0,0,0,0.15)',
          'fill-outline-color': 'rgba(0,0,0,0.25)',
        },
      })
      added.current = true
    }

    if (map.getLayer('shadows')) {
      map.setLayoutProperty('shadows', 'visibility', 'visible')
    }
  }, [map, date, lat, lng, buildings])

  return null
}

// Simple equirectangular projection for local shadow computation
function projectPoint(lng: number, lat: number, refLng: number, refLat: number): [number, number] {
  const latRad = refLat * Math.PI / 180
  const x = (lng - refLng) * Math.cos(latRad) * 111320
  const y = (lat - refLat) * 111320
  return [x, y]
}

function unprojectPoint(x: number, y: number, refLng: number, refLat: number): [number, number] {
  const latRad = refLat * Math.PI / 180
  const lng = refLng + x / (Math.cos(latRad) * 111320)
  const lat = refLat + y / 111320
  return [lng, lat]
}
