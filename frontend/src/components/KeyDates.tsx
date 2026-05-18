interface KeyDatesProps {
  bestDay: number
  worstDay: number
  onDaySelect: (day: number) => void
}

const SOLSTICE_EQUINOX = [
  { label: 'Spring Equinox', doy: 79 },
  { label: 'Summer Solstice', doy: 171 },
  { label: 'Autumn Equinox', doy: 265 },
  { label: 'Winter Solstice', doy: 355 },
]

export default function KeyDates({ bestDay, worstDay, onDaySelect }: KeyDatesProps) {
  return (
    <div style={{ marginBottom: 20 }}>
      <label style={{ fontSize: 12, fontWeight: 600, color: '#555', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
        Key Dates
      </label>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 4, marginTop: 4 }}>
        {SOLSTICE_EQUINOX.map((d) => (
          <button
            key={d.label}
            onClick={() => onDaySelect(d.doy)}
            style={dateBtnStyle}
          >
            {d.label} (Day {d.doy + 1})
          </button>
        ))}
        <button
          onClick={() => onDaySelect(bestDay)}
          style={{ ...dateBtnStyle, color: '#e67e22' }}
        >
          Sunniest Day (Day {bestDay + 1})
        </button>
        <button
          onClick={() => onDaySelect(worstDay)}
          style={{ ...dateBtnStyle, color: '#3498db' }}
        >
          Shadiest Day (Day {worstDay + 1})
        </button>
      </div>
    </div>
  )
}

const dateBtnStyle: React.CSSProperties = {
  textAlign: 'left',
  padding: '6px 10px',
  fontSize: 13,
  border: '1px solid #e0e0e0',
  borderRadius: 4,
  background: '#f9f9f9',
  cursor: 'pointer',
  color: '#555',
}
