import { useEffect, useRef, useState, useCallback } from 'react'
import maplibregl from 'maplibre-gl'

interface RectangleDrawProps {
  map: maplibregl.Map | null
  onRectangle: (bounds: { lat1: number; lng1: number; lat2: number; lng2: number }) => void
}

export default function RectangleDraw({ map, onRectangle }: RectangleDrawProps) {
  const [drawing, setDrawing] = useState(false)
  const startRef = useRef<maplibregl.LngLat | null>(null)
  const rectLayer = useRef<string | null>(null)

  const startDrawing = useCallback(() => {
    setDrawing(true)
    startRef.current = null
  }, [])

  useEffect(() => {
    if (!map) return
    if (!drawing) return

    const controller = new AbortController()
    const signal = controller.signal

    const onClick = (e: maplibregl.MapMouseEvent) => {
      if (signal.aborted) return

      if (!startRef.current) {
        startRef.current = e.lngLat
        return
      }

      const p1 = startRef.current
      const p2 = e.lngLat

      if (rectLayer.current && map.getLayer(rectLayer.current)) {
        map.removeLayer(rectLayer.current)
        map.removeSource('rect')
        rectLayer.current = null
      }

      const coords = [
        [p1.lng, p1.lat],
        [p2.lng, p1.lat],
        [p2.lng, p2.lat],
        [p1.lng, p2.lat],
        [p1.lng, p1.lat],
      ]

      map.addSource('rect', {
        type: 'geojson',
        data: {
          type: 'Feature',
          properties: {},
          geometry: { type: 'Polygon', coordinates: [coords] },
        },
      })
      map.addLayer({ id: 'rect-outline', type: 'line', source: 'rect', paint: { 'line-color': '#e74c3c', 'line-width': 2 } })
      map.addLayer({ id: 'rect-fill', type: 'fill', source: 'rect', paint: { 'fill-color': '#e74c3c', 'fill-opacity': 0.1 } })
      rectLayer.current = 'rect-outline'

      startRef.current = null
      setDrawing(false)

      onRectangle({ lat1: p1.lat, lng1: p1.lng, lat2: p2.lat, lng2: p2.lng })
    }

    map.on('click', onClick)

    map.getCanvas().style.cursor = 'crosshair'

    return () => {
      controller.abort()
      map.off('click', onClick)
      map.getCanvas().style.cursor = ''
    }
  }, [map, drawing, onRectangle])

  useEffect(() => {
    if (!map) return
    return () => {
      if (rectLayer.current && map.getLayer(rectLayer.current)) {
        try {
          map.removeLayer(rectLayer.current)
          map.removeSource('rect')
        } catch {}
      }
    }
  }, [map])

  return (
    <button
      onClick={(e) => { e.stopPropagation(); startDrawing() }}
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
