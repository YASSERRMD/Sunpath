import { useState } from 'react'

export default function EmbedPanel() {
  const [open, setOpen] = useState(false)

  const embedCode = `<iframe
  src="${window.location.origin}/?embed=1&lat=48.8566&lng=2.3522&h=1.5"
  width="100%"
  height="500"
  style="border: none; border-radius: 8px;"
  title="Sunpath solar analysis"
></iframe>`

  return (
    <div style={{ marginTop: 16, borderTop: '1px solid #e0e0e0', paddingTop: 16 }}>
      <button
        onClick={() => setOpen(!open)}
        style={{
          background: 'none', border: 'none', color: '#999', fontSize: 12,
          cursor: 'pointer', textDecoration: 'underline', padding: 0,
        }}
      >
        {open ? 'Hide' : 'Embed this map'}
      </button>

      {open && (
        <div style={{ marginTop: 8 }}>
          <p style={{ fontSize: 12, color: '#777', margin: '0 0 6px 0' }}>
            Copy this HTML to embed the current view on your site:
          </p>
          <textarea
            readOnly
            value={embedCode}
            onClick={(e) => (e.target as HTMLTextAreaElement).select()}
            style={{
              width: '100%', padding: 8, fontSize: 11, border: '1px solid #ddd',
              borderRadius: 4, fontFamily: 'monospace', resize: 'none',
              height: 120, background: '#f8f8f8', color: '#333',
            }}
          />
        </div>
      )}
    </div>
  )
}
