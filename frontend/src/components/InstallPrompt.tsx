import { useState, useEffect } from 'react'

export default function InstallPrompt() {
  const [deferredPrompt, setDeferredPrompt] = useState<Event | null>(null)
  const [show, setShow] = useState(false)

  useEffect(() => {
    const handler = (e: Event) => {
      e.preventDefault()
      setDeferredPrompt(e)
      setShow(true)
    }
    window.addEventListener('beforeinstallprompt', handler)
    return () => window.removeEventListener('beforeinstallprompt', handler)
  }, [])

  const install = () => {
    if (!deferredPrompt) return
    ;(deferredPrompt as any).prompt()
    ;(deferredPrompt as any).userChoice.then(() => {
      setDeferredPrompt(null)
      setShow(false)
    })
  }

  if (!show) return null

  return (
    <div style={{ position: 'fixed', bottom: 16, left: 16, right: 16, zIndex: 1000, maxWidth: 400, margin: '0 auto' }}>
      <div style={{ background: '#fff', border: '1px solid #ddd', borderRadius: 12, padding: '12px 16px', boxShadow: '0 4px 12px rgba(0,0,0,0.15)', display: 'flex', alignItems: 'center', gap: 12 }}>
        <div style={{ flex: 1, fontSize: 13, lineHeight: 1.4 }}>
          Install <strong>Sunpath</strong> for offline access
        </div>
        <button onClick={install} style={{ background: '#2ecc71', color: '#fff', border: 'none', borderRadius: 8, padding: '8px 16px', fontSize: 13, fontWeight: 600, cursor: 'pointer' }}>
          Install
        </button>
        <button onClick={() => setShow(false)} style={{ background: 'none', border: 'none', color: '#999', cursor: 'pointer', fontSize: 18, lineHeight: 1, padding: 4 }}>
          &times;
        </button>
      </div>
    </div>
  )
}
