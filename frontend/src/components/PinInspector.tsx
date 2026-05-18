import { useMemo } from 'react'
import type { DayResult } from '../lib/horizon'

interface PinInspectorProps {
  dayResult: DayResult | null
  selectedDay: number
  onDayChange: (day: number) => void
  days: DayResult[]
}

export default function PinInspector({ dayResult, selectedDay, onDayChange, days }: PinInspectorProps) {
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
                  flex: 1,
                  height: Math.max(h, 1),
                  background: i === selectedDay ? '#e74c3c' : 'hsl(35, 80%, 45%)',
                  borderRadius: '1px 1px 0 0',
                  cursor: 'pointer',
                  opacity: i === selectedDay ? 1 : 0.6,
                }}
              />
            )
          })}
        </div>
      </div>

      <div style={{ marginBottom: 20 }}>
        <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          Selected Day
        </label>
        <p style={{ fontSize: 14, margin: '4px 0 0 0' }}>
          Day {selectedDay + 1} of 365
        </p>
        <p style={{ fontSize: 24, fontWeight: 700, margin: '4px 0', color: '#333' }}>
          {summary}
        </p>
        {dayResult && (
          <p style={{ fontSize: 13, color: '#666', margin: 0 }}>
            First sun: {formatMinute(dayResult)} | Last sun: {formatLastSun(dayResult)}
          </p>
        )}
      </div>

      <div style={{ marginBottom: 20 }}>
        <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          Date Scrubber
        </label>
        <input
          type="range"
          min={0}
          max={364}
          value={selectedDay}
          onChange={(e) => onDayChange(parseInt(e.target.value))}
          aria-label={`Select day of year. Currently day ${selectedDay + 1}`}
          style={{ width: '100%', marginTop: 4 }}
        />
        <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: '#999' }}>
          <span>Jan 1</span>
          <span>Dec 31</span>
        </div>
      </div>
    </div>
  )
}

function formatMinute(day: DayResult): string {
  const sun = day.sunStates.find(s => s.inSun)
  if (!sun) return '--:--'
  return sun.time.toISOString().slice(11, 16)
}

function formatLastSun(day: DayResult): string {
  let last: { time: Date; inSun: boolean } | null = null
  for (const s of day.sunStates) {
    if (s.inSun) last = s
  }
  if (!last) return '--:--'
  return last.time.toISOString().slice(11, 16)
}
