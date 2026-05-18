import { useState, useCallback, useEffect } from 'react'
import MapView from './components/MapView'
import SidePanel from './components/SidePanel'
import YearHeatmap from './components/YearHeatmap'
import PinInspector from './components/PinInspector'
import KeyDates from './components/KeyDates'
import ConfidenceBanner from './components/ConfidenceBanner'
import AboutPanel from './components/AboutPanel'
import { fetchHorizon } from './lib/api'
import { computeYear } from './lib/horizon'
import { generateSummary } from './lib/summary'
import { decodeState, updateURL } from './lib/urlstate'
import type { HorizonProfile } from './lib/api'
import type { YearResult, DayResult } from './lib/horizon'

export interface PinState {
  lat: number
  lng: number
}

type LoadState = 'idle' | 'loading' | 'loaded' | 'error'

function App() {
  const [pin, setPin] = useState<PinState | null>(null)
  const [height, setHeight] = useState(1.5)
  const [profile, setProfile] = useState<HorizonProfile | null>(null)
  const [year, setYear] = useState<YearResult | null>(null)
  const [loadState, setLoadState] = useState<LoadState>('idle')
  const [loadError, setLoadError] = useState('')
  const [selectedDay, setSelectedDay] = useState(170)
  const [offline, setOffline] = useState(!navigator.onLine)

  useEffect(() => {
    const handleOffline = () => setOffline(true)
    const handleOnline = () => setOffline(false)
    window.addEventListener('offline', handleOffline)
    window.addEventListener('online', handleOnline)
    return () => {
      window.removeEventListener('offline', handleOffline)
      window.removeEventListener('online', handleOnline)
    }
  }, [])

  useEffect(() => {
    const state = decodeState()
    if (state) {
      setPin({ lat: state.lat, lng: state.lng })
      setHeight(state.h)
    }
  }, [])

  useEffect(() => {
    if (pin) {
      updateURL(pin.lat, pin.lng, height)
    }
  }, [pin, height])

  const handlePinChange = useCallback((p: PinState) => {
    setPin(p)
    setProfile(null)
    setYear(null)
    setLoadState('idle')
  }, [])

  useEffect(() => {
    if (!pin || offline) return
    setLoadState('loading')
    setLoadError('')

    fetchHorizon(pin.lat, pin.lng, height)
      .then((p) => {
        setProfile(p)
        const y = computeYear(pin.lat, pin.lng, p)
        setYear(y)
        setLoadState('loaded')
      })
      .catch((err) => {
        const msg = err.message || ''
        if (msg.includes('502') || msg.includes('fetch')) {
          setLoadError('Building data is too thin in this area for a reliable analysis.')
        } else {
          setLoadError(msg || 'Failed to compute. Please try again.')
        }
        setLoadState('error')
      })
  }, [pin, height, offline])

  const dayResult: DayResult | null = year && selectedDay >= 0 && selectedDay < year.days.length
    ? year.days[selectedDay]
    : null

  const summary = year && pin
    ? generateSummary(year, pin.lat, pin.lng, height)
    : ''

  return (
    <div style={{ display: 'flex', width: '100vw', height: '100vh', overflow: 'hidden', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <MapView pin={pin} onPinChange={handlePinChange} />
      </div>
      <div style={{
        width: 400,
        background: '#fff',
        borderLeft: '1px solid #e0e0e0',
        padding: 24,
        overflowY: 'auto',
      }}>
        <h1 style={{ fontSize: 20, fontWeight: 600, margin: '0 0 4px 0' }}>Sunpath</h1>
        <p style={{ fontSize: 13, color: '#666', margin: '0 0 24px 0' }}>
          Solar exposure analysis for any point
        </p>

        <SidePanel pin={pin} height={height} onHeightChange={setHeight} onPinChange={handlePinChange} />

        {offline && (
          <div style={{ padding: '10px 14px', background: '#fff3e0', border: '1px solid #ffcc80', borderRadius: 6, marginBottom: 16, fontSize: 13, color: '#e65100' }}>
            You are offline. Sunpath needs a network connection to fetch building data.
          </div>
        )}

        {loadState === 'loading' && (
          <div style={{ padding: '20px 0', textAlign: 'center', color: '#999', fontSize: 14 }}>
            Fetching building data and computing horizon...
          </div>
        )}

        {loadState === 'error' && (
          <div style={{ padding: '20px 0', textAlign: 'center' }}>
            <p style={{ color: '#e74c3c', fontSize: 14 }}>{loadError}</p>
            <button
              onClick={() => setLoadState('idle')}
              style={{ marginTop: 8, padding: '6px 16px', fontSize: 13, border: '1px solid #ccc', borderRadius: 4, background: '#fff', cursor: 'pointer' }}
            >
              Try again
            </button>
          </div>
        )}

        {loadState === 'loaded' && profile && year && (
          <>
            {profile.confidence < 0.7 && (
              <ConfidenceBanner
                confidence={profile.confidence}
                buildingCount={profile.building_count}
                estimatedCount={profile.estimated_count}
              />
            )}

            <YearHeatmap
              grid={year.grid}
              selectedDay={selectedDay}
              onDaySelect={setSelectedDay}
            />

            <PinInspector
              dayResult={dayResult}
              selectedDay={selectedDay}
              onDayChange={setSelectedDay}
              days={year.days}
            />

            <KeyDates
              bestDay={year.bestDay}
              worstDay={year.worstDay}
              onDaySelect={setSelectedDay}
            />

            <div style={{
              padding: '12px 14px',
              background: '#f0f8ff',
              border: '1px solid #b3d9ff',
              borderRadius: 6,
              fontSize: 13,
              lineHeight: 1.5,
              color: '#2c3e50',
            }}>
              {summary}
            </div>
          </>
        )}
        <AboutPanel />
      </div>
    </div>
  )
}

export default App
