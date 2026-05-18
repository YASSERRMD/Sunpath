import { useEffect, useRef } from 'react'
import maplibregl from 'maplibre-gl'

interface GridCell {
  lat: number
  lng: number
  sun_minutes: number
}

interface GridHeatmapProps {
  map: maplibregl.Map | null
  cells: GridCell[]
  onCellClick?: (cell: GridCell) => void
}

export default function GridHeatmap({ map, cells, onCellClick }: GridHeatmapProps) {
  const added = useRef(false)

  useEffect(() => {
    if (!map || cells.length === 0) return

    const maxMin = Math.max(...cells.map(c => c.sun_minutes), 1)

    const features: GeoJSON.Feature[] = cells.map((c) => ({
      type: 'Feature',
      properties: {
        sun_minutes: c.sun_minutes,
        intensity: c.sun_minutes / maxMin,
      },
      geometry: {
        type: 'Point',
        coordinates: [c.lng, c.lat],
      },
    }))

    const source = map.getSource('grid') as maplibregl.GeoJSONSource
    if (source) {
      source.setData({ type: 'FeatureCollection', features })
    } else {
      map.addSource('grid', {
        type: 'geojson',
        data: { type: 'FeatureCollection', features },
      })
      map.addLayer({
        id: 'grid-heat',
        type: 'heatmap',
        source: 'grid',
        paint: {
          'heatmap-weight': ['get', 'intensity'],
          'heatmap-intensity': 1,
          'heatmap-radius': 20,
          'heatmap-color': [
            'interpolate',
            ['linear'],
            ['heatmap-density'],
            0, 'rgba(0,0,255,0)',
            0.2, 'rgba(100,100,255,0.5)',
            0.4, 'rgba(100,200,100,0.6)',
            0.6, 'rgba(255,200,50,0.7)',
            0.8, 'rgba(255,100,50,0.8)',
            1, 'rgba(200,50,0,0.9)',
          ],
          'heatmap-opacity': 0.8,
        },
      })
      added.current = true
    }

    if (onCellClick) {
      map.on('click', 'grid-heat', (e) => {
        if (e.features && e.features[0]) {
          const props = e.features[0].properties
          onCellClick({
            lat: props!.lat,
            lng: props!.lng,
            sun_minutes: props!.sun_minutes,
          })
        }
      })
    }

    return () => {
      if (added.current && map.getLayer('grid-heat')) {
        map.removeLayer('grid-heat')
        map.removeSource('grid')
        added.current = false
      }
    }
  }, [map, cells, onCellClick])

  if (cells.length === 0) return null

  const best = cells.reduce((a, b) => a.sun_minutes > b.sun_minutes ? a : b)
  const worst = cells.reduce((a, b) => a.sun_minutes < b.sun_minutes ? a : b)

  return (
    <div style={{ position: 'absolute', bottom: 60, left: 16, zIndex: 10, background: '#fff', borderRadius: 6, padding: '8px 12px', fontSize: 12, boxShadow: '0 2px 8px rgba(0,0,0,0.15)' }}>
      <div style={{ color: '#e67e22' }}>Sunniest: {best.sun_minutes} min</div>
      <div style={{ color: '#3498db' }}>Shadiest: {worst.sun_minutes} min</div>
    </div>
  )
}
