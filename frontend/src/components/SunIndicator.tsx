import type { HorizonProfile } from '../lib/api'
import { isInDirectSun } from '../lib/horizon'
import { getSunPosition } from '../lib/sun'

interface SunIndicatorProps {
  date: Date
  lat: number
  lng: number
  profile: HorizonProfile
}

export default function SunIndicator({ date, lat, lng, profile }: SunIndicatorProps) {
  const inSun = isInDirectSun(date, lat, lng, profile)
  const { elevation } = getSunPosition(date, lat, lng)

  return (
    <div style={{
      display: 'flex',
      alignItems: 'center',
      gap: 8,
      padding: '8px 12px',
      background: inSun ? '#fff8e1' : '#f0f0f0',
      border: `1px solid ${inSun ? '#ffb300' : '#ccc'}`,
      borderRadius: 6,
      marginBottom: 16,
    }}>
      <div style={{
        width: 16,
        height: 16,
        borderRadius: '50%',
        background: inSun ? '#ffb300' : '#999',
        boxShadow: inSun ? '0 0 8px rgba(255,179,0,0.5)' : 'none',
        flexShrink: 0,
      }} />
      <div>
        <div style={{ fontSize: 14, fontWeight: 600, color: inSun ? '#e65100' : '#555' }}>
          {inSun ? 'In direct sun' : 'In shade'}
        </div>
        <div style={{ fontSize: 12, color: '#888' }}>
          Sun elevation: {elevation.toFixed(1)} degrees
        </div>
      </div>
    </div>
  )
}
