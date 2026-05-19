import { useMemo } from 'react'
import type { DayResult } from '../lib/horizon'
import { formatMinutes } from '../lib/timezone'

interface PinInspectorProps {
  dayResult: DayResult | null
  selectedDay: number
  onDayChange: (day: number) => void
  days: DayResult[]
  timezone?: string
}

export default function PinInspector({ dayResult, selectedDay, onDayChange, days, timezone }: PinInspectorProps) {
  const tz = timezone || 'UTC+00:00'
  const summary = useMemo(() => {
    if (!dayResult) return ''
    const h = Math.floor(dayResult.totalMinutes / 60)
    const m = dayResult.totalMinutes % 60
    const parts: string[] = []
    if (h > 0) parts.push(`${h}h`)
    if (m > 0) parts.push(`${m}m`)
    return parts.join(' ') || '0m'
  }, [dayResult])

  return (
    <div>
      <div style={{ marginBottom: 12, fontSize: 16, fontWeight: 600 }}>
        {summary}
      </div>
      <div style={{ marginBottom: 20 }}>
        <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          Daily Sun Hours
        </label>
        <div style={{ display: 'flex', gap: 2, marginTop: 4, height: 40, alignItems: 'flex-end' }}>
          {days.slice(0, 365).map((d, i) => {
            const maxMin = Math.max(...days.slice(0, 365).map(x => x.totalMinutes), 1)
            const h = (d.totalMinutes / maxMin) * 36
            return (
              <div
                key={i}
                onClick={() => onDayChange(i)}
                style={{
                  flex: 1, height: '100%', display: 'flex', flexDirection: 'column', justifyContent: 'flex-end',
                  cursor: 'pointer', position: 'relative',
                }}
              >
                <div style={{
                  height: Math.max(h, 1), background: d.totalMinutes > 0 ? '#f39c12' : '#ddd',
                  borderRadius: '1px 1px 0 0', opacity: i === selectedDay ? 1 : 0.6,
                }} />
              </div>
            )
          })}
        </div>
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: '#999' }}>
          <span>Jan 1</span>
          <span>Dec 31</span>
        </div>
      </div>

      {dayResult && dayResult.totalMinutes > 0 && (
        <div style={{ fontSize: 13, lineHeight: 1.6, color: '#2c3e50', marginBottom: 16, padding: '10px 14px', background: '#fef9e7', borderRadius: 6, border: '1px solid #f9e79f' }}>
          <strong>First sun: {formatFirstSun(dayResult, tz)} &middot; Last sun: {formatLastSun(dayResult, tz)}</strong>
          <div style={{ fontSize: 11, color: '#999', marginTop: 2 }}>
            Times shown in {tz}
          </div>
        </div>
      )}
    </div>
  )
}

function formatFirstSun(day: DayResult, tz: string): string {
  if (day.firstSunMinute < 0) return '--:--'
  return formatMinutes(day.firstSunMinute, tz)
}

function formatLastSun(day: DayResult, tz: string): string {
  if (day.lastSunMinute < 0) return '--:--'
  return formatMinutes(day.lastSunMinute, tz)
}
