import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'
import 'maplibre-gl/dist/maplibre-gl.css'
import type { PinState } from '../App'
import ShadowOverlay from './ShadowOverlay'
import RectangleDraw from './RectangleDraw'
import GridHeatmap from './GridHeatmap'
import type { BuildingOutline, GridCell } from '../lib/api'

interface MapViewProps {
  pin: PinState | null
  onPinChange: (p: PinState) => void
  sweepDate?: Date
  buildings?: BuildingOutline[]
  gridCells?: GridCell[]
  gridLoading?: boolean
  onRectangle?: (bounds: { lat1: number; lng1: number; lat2: number; lng2: number }) => void
  onGridCellClick?: (cell: GridCell) => void
}

const TILE_STYLE_URL = import.meta.env.VITE_TILE_STYLE_URL || 'https://demotiles.maplibre.org/style.json'

export default function MapView({ pin, onPinChange, sweepDate, buildings, gridCells, gridLoading, onRectangle, onGridCellClick }: MapViewProps) {
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

  return (
    <div style={{ width: '100%', height: '100%', position: 'relative' }}>
      <div ref={containerRef} style={{ width: '100%', height: '100%' }} />
      {pin && sweepDate && buildings && (
        <ShadowOverlay
          map={mapRef.current}
          date={sweepDate}
          lat={pin.lat}
          lng={pin.lng}
          buildings={buildings}
        />
      )}
      {onRectangle && (
        <RectangleDraw map={mapRef.current} onRectangle={onRectangle} />
      )}
      {gridCells && (
        <GridHeatmap map={mapRef.current} cells={gridCells} onCellClick={onGridCellClick} />
      )}
      {gridLoading && (
        <div style={{ position: 'absolute', top: 16, right: 16, background: '#fff', padding: '6px 12px', borderRadius: 4, fontSize: 12, boxShadow: '0 1px 4px rgba(0,0,0,0.15)' }}>
          Loading grid...
        </div>
      )}
    </div>
  )
}
