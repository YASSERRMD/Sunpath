import type { PinState } from '../App'
import GeocodeSearch from './GeocodeSearch'

interface SidePanelProps {
  pin: PinState | null
  height: number
  onHeightChange: (h: number) => void
  onPinChange: (p: PinState) => void
}

export default function SidePanel({ pin, height, onHeightChange, onPinChange }: SidePanelProps) {
  return (
    <div style={{
      width: 360,
      background: '#fff',
      borderLeft: '1px solid #e0e0e0',
      padding: 24,
      overflowY: 'auto',
      fontFamily: 'system-ui, -apple-system, sans-serif',
    }}>
      <h1 style={{ fontSize: 20, fontWeight: 600, margin: '0 0 4px 0' }}>Sunpath</h1>
      <p style={{ fontSize: 13, color: '#666', margin: '0 0 24px 0' }}>
        Solar exposure analysis for any point
      </p>

      <GeocodeSearch onSelect={onPinChange} />

      {!pin && (
        <p style={{ color: '#999', fontSize: 14 }}>
          Click the map to drop a pin and start exploring.
        </p>
      )}

      {pin && (
        <>
          <div style={{ marginBottom: 20 }}>
            <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Coordinates
            </label>
            <p style={{ fontSize: 14, margin: '4px 0 0 0', fontFamily: 'monospace' }}>
              {pin.lat.toFixed(6)}, {pin.lng.toFixed(6)}
            </p>
          </div>

          <div style={{ marginBottom: 20 }}>
            <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
              Observer Height
            </label>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 4 }}>
              <input
                type="range"
                min={0}
                max={100}
                step={0.5}
                value={height}
                onChange={(e) => onHeightChange(parseFloat(e.target.value))}
                style={{ flex: 1 }}
              />
              <span style={{ fontSize: 14, fontFamily: 'monospace', minWidth: 60 }}>
                {height.toFixed(1)}m
              </span>
            </div>
            <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
              <button
                onClick={() => onHeightChange(1.5)}
                style={btnStyle(height === 1.5)}
              >
                Ground
              </button>
              <button
                onClick={() => onHeightChange(10)}
                style={btnStyle(height === 10)}
              >
                Floor 3
              </button>
              <button
                onClick={() => onHeightChange(30)}
                style={btnStyle(height === 30)}
              >
                Floor 10
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}

function btnStyle(active: boolean): React.CSSProperties {
  return {
    padding: '4px 12px',
    fontSize: 12,
    border: active ? '2px solid #3498db' : '1px solid #ccc',
    borderRadius: 4,
    background: active ? '#ebf5fb' : '#f8f8f8',
    cursor: 'pointer',
    color: active ? '#2980b9' : '#555',
    fontWeight: active ? 600 : 400,
  }
}
