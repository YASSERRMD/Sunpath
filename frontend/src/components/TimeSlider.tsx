import { useState, useRef, useCallback, useEffect } from 'react'
import { formatMinutes } from '../lib/timezone'

interface TimeSliderProps {
  selectedDay: number
  onTimeChange: (hourFraction: number) => void
  timezone?: string
}

export default function TimeSlider({ selectedDay: _selectedDay, onTimeChange, timezone }: TimeSliderProps) {
  const tz = timezone || 'UTC+00:00'
  const [timeFrac, setTimeFrac] = useState(0.5)
  const [playing, setPlaying] = useState(false)
  const animRef = useRef<ReturnType<typeof requestAnimationFrame>>()
  const lastRef = useRef<number>(0)

  const frame = useCallback((timestamp: number) => {
    if (!lastRef.current) lastRef.current = timestamp
    const delta = (timestamp - lastRef.current) / 1000
    lastRef.current = timestamp

    setTimeFrac((prev) => {
      const next = prev + delta / 120
      return next > 1 ? 0 : next
    })

    animRef.current = requestAnimationFrame(frame)
  }, [])

  useEffect(() => {
    if (playing) {
      lastRef.current = 0
      animRef.current = requestAnimationFrame(frame)
    } else {
      if (animRef.current) cancelAnimationFrame(animRef.current)
    }
    return () => {
      if (animRef.current) cancelAnimationFrame(animRef.current)
    }
  }, [playing, frame])

  useEffect(() => {
    onTimeChange(timeFrac)
  }, [timeFrac, onTimeChange])

  const totalMinutes = Math.floor(timeFrac * 1440)

  return (
    <div style={{ marginBottom: 20 }}>
      <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
        Time of Day
      </label>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 4 }}>
        <button
          onClick={() => setPlaying(!playing)}
          style={{
            padding: '6px 12px',
            fontSize: 14,
            border: '1px solid #ccc',
            borderRadius: 4,
            background: playing ? '#e74c3c' : '#3498db',
            color: '#fff',
            cursor: 'pointer',
            minWidth: 60,
          }}
        >
          {playing ? 'Stop' : 'Play'}
        </button>
        <input
          type="range"
          min={0}
          max={1}
          step={0.001}
          value={timeFrac}
          onChange={(e) => setTimeFrac(parseFloat(e.target.value))}
          style={{ flex: 1 }}
          aria-label="Time of day slider"
        />
        <span style={{ fontSize: 14, fontFamily: 'monospace', minWidth: 60, textAlign: 'right' }}>
          {formatMinutes(totalMinutes, tz)}
        </span>
      </div>
    </div>
  )
}
