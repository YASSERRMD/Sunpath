import { useState, useCallback, useRef } from 'react'
import type { PinState } from '../App'
import { fetchGeocode } from '../lib/api'
import type { GeocodeResult } from '../lib/api'

interface GeocodeSearchProps {
  onSelect: (pin: PinState) => void
}

export default function GeocodeSearch({ onSelect }: GeocodeSearchProps) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<GeocodeResult[]>([])
  const [loading, setLoading] = useState(false)
  const [showResults, setShowResults] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout>>()

  const search = useCallback(async (q: string) => {
    if (q.length < 2) {
      setResults([])
      return
    }
    setLoading(true)
    try {
      const data = await fetchGeocode(q)
      setResults(data)
      setShowResults(true)
    } catch {
      setResults([])
    } finally {
      setLoading(false)
    }
  }, [])

  const handleInput = useCallback((value: string) => {
    setQuery(value)
    if (timerRef.current) clearTimeout(timerRef.current)
    timerRef.current = setTimeout(() => search(value), 300)
  }, [search])

  const handleSelect = useCallback((r: GeocodeResult) => {
    onSelect({ lat: parseFloat(r.lat), lng: parseFloat(r.lon) })
    setQuery(r.display_name)
    setShowResults(false)
  }, [onSelect])

  return (
    <div style={{ position: 'relative', marginBottom: 20 }}>
      <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em', display: 'block', marginBottom: 4 }}>
        Search Location
      </label>
      <input
        type="text"
        value={query}
        onChange={(e) => handleInput(e.target.value)}
        onFocus={() => results.length > 0 && setShowResults(true)}
        onBlur={() => setTimeout(() => setShowResults(false), 200)}
        placeholder="e.g. Paris, London, Tokyo..."
        aria-label="Search for a location"
        style={{
          width: '100%',
          padding: '8px 12px',
          fontSize: 14,
          border: '1px solid #ccc',
          borderRadius: 4,
          boxSizing: 'border-box',
        }}
      />
      {loading && (
        <span style={{ position: 'absolute', right: 8, top: 32, fontSize: 12, color: '#999' }}>
          Searching...
        </span>
      )}
      {showResults && results.length > 0 && (
        <ul style={{
          position: 'absolute',
          top: '100%',
          left: 0,
          right: 0,
          background: '#fff',
          border: '1px solid #e0e0e0',
          borderRadius: 4,
          listStyle: 'none',
          margin: '4px 0 0 0',
          padding: 0,
          maxHeight: 240,
          overflowY: 'auto',
          zIndex: 100,
          boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
        }}>
          {results.map((r, i) => (
            <li
              key={i}
              onMouseDown={() => handleSelect(r)}
              style={{
                padding: '8px 12px',
                cursor: 'pointer',
                fontSize: 13,
                borderBottom: i < results.length - 1 ? '1px solid #f0f0f0' : 'none',
              }}
              onMouseEnter={(e) => (e.currentTarget.style.background = '#f5f5f5')}
              onMouseLeave={(e) => (e.currentTarget.style.background = '')}
            >
              {r.display_name}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
