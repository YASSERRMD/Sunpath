import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import type { PinState } from '../App'

interface MapViewProps {
  pin: PinState | null
  onPinChange: (p: PinState) => void
}

const TILE_STYLE_URL = import.meta.env.VITE_TILE_STYLE_URL || 'https://demotiles.maplibre.org/style.json'

export default function MapView({ pin, onPinChange }: MapViewProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const mapRef = useRef<maplibregl.Map | null>(null)
  const markerRef = useRef<maplibregl.Marker | null>(null)

  useEffect(() => {
    if (!containerRef.current || mapRef.current) return

    const map = new maplibregl.Map({
      container: containerRef.current,
      style: TILE_STYLE_URL,
      center: [0, 20],
      zoom: 2,
    })
    map.addControl(new maplibregl.NavigationControl(), 'top-right')
    mapRef.current = map

    map.on('click', (e) => {
      onPinChange({ lat: e.lngLat.lat, lng: e.lngLat.lng })
    })

    return () => {
      map.remove()
      mapRef.current = null
    }
  }, [onPinChange])

  useEffect(() => {
    if (!mapRef.current) return
    const map = mapRef.current

    if (!pin) {
      if (markerRef.current) {
        markerRef.current.remove()
        markerRef.current = null
      }
      return
    }

    if (markerRef.current) {
      markerRef.current.setLngLat([pin.lng, pin.lat])
    } else {
      const el = document.createElement('div')
      el.style.width = '24px'
      el.style.height = '24px'
      el.style.borderRadius = '50%'
      el.style.background = '#e74c3c'
      el.style.border = '3px solid white'
      el.style.cursor = 'grab'
      el.style.boxShadow = '0 2px 6px rgba(0,0,0,0.3)'

      const marker = new maplibregl.Marker({ element: el, draggable: true })
        .setLngLat([pin.lng, pin.lat])
        .addTo(map)

      marker.on('dragend', () => {
        const lngLat = marker.getLngLat()
        onPinChange({ lat: lngLat.lat, lng: lngLat.lng })
      })

      markerRef.current = marker
    }

    map.flyTo({ center: [pin.lng, pin.lat], zoom: 15 })
  }, [pin, onPinChange])

  return <div ref={containerRef} style={{ width: '100%', height: '100%' }} />
}
