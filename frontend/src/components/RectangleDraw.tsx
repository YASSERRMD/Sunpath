import { useEffect, useRef, useState, useCallback } from 'react'
import maplibregl from 'maplibre-gl'

interface RectangleDrawProps {
  map: maplibregl.Map | null
  onRectangle: (bounds: { lat1: number; lng1: number; lat2: number; lng2: number }) => void
}

export default function RectangleDraw({ map, onRectangle }: RectangleDrawProps) {
  const [drawing, setDrawing] = useState(false)
  const startRef = useRef<{ lat: number; lng: number } | null>(null)
  const rectRef = useRef<maplibregl.GeoJSONSource | null>(null)

  useEffect(() => {
    if (!map) return

    map.on('click', (e) => {
      if (!drawing) return

      if (!startRef.current) {
        startRef.current = { lat: e.lngLat.lat, lng: e.lngLat.lng }
        return
      }

      const p1 = startRef.current
      const p2 = { lat: e.lngLat.lat, lng: e.lngLat.lng }
      startRef.current = null
      setDrawing(false)

      onRectangle({
        lat1: p1.lat,
        lng1: p1.lng,
        lat2: p2.lat,
        lng2: p2.lng,
      })
    })
  }, [map, drawing, onRectangle])

  const startDrawing = useCallback(() => {
    setDrawing(true)
    startRef.current = null
  }, [])

  return (
    <button
      onClick={startDrawing}
      style={{
        position: 'absolute',
        bottom: 16,
        left: 16,
        zIndex: 10,
        padding: '8px 16px',
        fontSize: 13,
        border: drawing ? '2px solid #e74c3c' : '1px solid #ccc',
        borderRadius: 4,
        background: drawing ? '#fff5f5' : '#fff',
        cursor: 'pointer',
        color: drawing ? '#e74c3c' : '#333',
        fontWeight: drawing ? 600 : 400,
      }}
    >
      {drawing ? 'Click two corners' : 'Explore Area'}
    </button>
  )
}
