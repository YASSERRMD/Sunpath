import type { PinState } from '../App'
import type { HorizonProfile } from '../lib/api'

interface ComparisonPanelProps {
  pins: PinState[]
  results: HorizonProfile[]
  loading: boolean
  height: number
  onAdd: () => void
  onRemove: (index: number) => void
  onRun: () => void
  onNavigate: (lat: number, lng: number) => void
}

export default function ComparisonPanel({ pins, results, loading, onAdd, onRemove, onRun, onNavigate }: ComparisonPanelProps) {
  if (pins.length === 0 && results.length === 0) {
    return (
      <div style={{ marginTop: 16, borderTop: '1px solid #eee', paddingTop: 16 }}>
        <button
          onClick={onAdd}
          style={{
            width: '100%', padding: '8px 12px', fontSize: 13,
            border: '1px dashed #3498db', borderRadius: 4,
            background: '#f0f8ff', color: '#3498db', cursor: 'pointer',
          }}
        >
          + Add current pin to compare
        </button>
      </div>
    )
  }

  return (
    <div style={{ marginTop: 16, borderTop: '1px solid #eee', paddingTop: 16 }}>
      <div style={{ fontSize: 13, fontWeight: 600, marginBottom: 8, color: '#2c3e50' }}>
        Multi-point comparison ({pins.length} pins)
      </div>

      {pins.map((p, i) => (
        <div key={`${p.lat}-${p.lng}`} style={{
          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
          padding: '4px 0', fontSize: 12, borderBottom: '1px solid #f0f0f0',
        }}>
          <div style={{ flex: 1, cursor: 'pointer' }} onClick={() => onNavigate(p.lat, p.lng)}>
            <span style={{ fontWeight: 500 }}>#{i + 1}</span> {p.lat.toFixed(4)}, {p.lng.toFixed(4)}
            {results[i] && (
              <span style={{ color: '#666', marginLeft: 6 }}>
                {results[i].confidence > 0 ? `${(results[i].confidence * 100).toFixed(0)}% conf` : ''}
              </span>
            )}
          </div>
          <button onClick={() => onRemove(i)} style={{ border: 'none', background: 'none', color: '#e74c3c', cursor: 'pointer', fontSize: 13, padding: '0 4px' }}>
            x
          </button>
        </div>
      ))}

      <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
        <button onClick={onAdd} style={{
          flex: 1, padding: '6px 8px', fontSize: 12,
          border: '1px dashed #3498db', borderRadius: 4,
          background: '#f0f8ff', color: '#3498db', cursor: 'pointer',
        }}>
          + Add current
        </button>
        {pins.length >= 2 && (
          <button onClick={onRun} disabled={loading} style={{
            flex: 1, padding: '6px 8px', fontSize: 12,
            border: '1px solid #27ae60', borderRadius: 4,
            background: '#27ae60', color: '#fff', cursor: loading ? 'default' : 'pointer',
            opacity: loading ? 0.6 : 1,
          }}>
            {loading ? 'Computing...' : 'Compare'}
          </button>
        )}
      </div>

      {results.length >= 2 && (
        <div style={{ marginTop: 12, fontSize: 12 }}>
          <div style={{ fontWeight: 600, marginBottom: 6, color: '#2c3e50' }}>Results</div>
          <div style={{ display: 'grid', gridTemplateColumns: `auto ${'1fr '.repeat(results.length)}`, gap: '4px 8px', alignItems: 'center' }}>
            <div style={{ fontWeight: 500, color: '#555' }}></div>
            {results.map((_, i) => (
              <div key={i} style={{ fontWeight: 500, textAlign: 'center', fontSize: 11, color: '#555' }}>
                #{i + 1}
              </div>
            ))}
            <div style={{ color: '#666' }}>Lat</div>
            {results.map((r, i) => (
              <div key={i} style={{ textAlign: 'center' }}>{r.lat.toFixed(4)}</div>
            ))}
            <div style={{ color: '#666' }}>Lng</div>
            {results.map((r, i) => (
              <div key={i} style={{ textAlign: 'center' }}>{r.lng.toFixed(4)}</div>
            ))}
            <div style={{ color: '#666' }}>Confidence</div>
            {results.map((r, i) => (
              <div key={i} style={{ textAlign: 'center', color: r.confidence < 0.7 ? '#e67e22' : '#27ae60' }}>
                {(r.confidence * 100).toFixed(0)}%
              </div>
            ))}
            <div style={{ color: '#666' }}>Buildings</div>
            {results.map((r, i) => (
              <div key={i} style={{ textAlign: 'center' }}>{r.building_count}</div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
