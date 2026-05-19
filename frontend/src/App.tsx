import { useState, useCallback, useEffect } from 'react'
import MapView from './components/MapView'
import SidePanel from './components/SidePanel'
import YearHeatmap from './components/YearHeatmap'
import PinInspector from './components/PinInspector'
import KeyDates from './components/KeyDates'
import ConfidenceBanner from './components/ConfidenceBanner'
import AboutPanel from './components/AboutPanel'
import EmbedPanel from './components/EmbedPanel'
import ProjectsPanel from './components/ProjectsPanel'
import ComparisonPanel from './components/ComparisonPanel'
import TimeSlider from './components/TimeSlider'
import InstallPrompt from './components/InstallPrompt'
import SunIndicator from './components/SunIndicator'
import { fetchHorizon, fetchBuildings, fetchGrid, type HorizonProfile, type BuildingOutline, type GridCell, fetchBatchHorizon } from './lib/api'
import { generateSummary } from './lib/summary'
import { resolveTimezone } from './lib/timezone'
import { decodeState, updateURL } from './lib/urlstate'
import { useWorker } from './lib/useWorker'

interface WorkerDayResult {
  date: string
  dayOfYear: number
  totalMinutes: number
  firstSunMinute: number
  lastSunMinute: number
}

interface WorkerYearResult {
  days: WorkerDayResult[]
  grid: number[][]
  maxSunMinutes: number
  minSunMinutes: number
  bestDay: number
  worstDay: number
}

interface DayResult {
  date: Date
  dayOfYear: number
  totalMinutes: number
  firstSunMinute: number
  lastSunMinute: number
  sunStates: { time: Date; inSun: boolean }[]
}

interface YearResult {
  days: DayResult[]
  grid: number[][]
  maxSunMinutes: number
  minSunMinutes: number
  bestDay: number
  worstDay: number
}

export interface PinState {
  lat: number
  lng: number
}

type LoadState = 'idle' | 'loading' | 'loaded' | 'error'

function App() {
  const [embed] = useState(() => new URLSearchParams(window.location.search).get('embed') === '1')
  const [pin, setPin] = useState<PinState | null>(null)
  const [height, setHeight] = useState(1.5)
  const [profile, setProfile] = useState<HorizonProfile | null>(null)
  const [year, setYear] = useState<YearResult | null>(null)
  const [buildings, setBuildings] = useState<BuildingOutline[]>([])
  const [loadState, setLoadState] = useState<LoadState>('idle')
  const [loadError, setLoadError] = useState('')
  const [selectedDay, setSelectedDay] = useState(170)
  const [timeFrac, setTimeFrac] = useState(0.5)
  const [offline, setOffline] = useState(!navigator.onLine)
  const [useDSM, setUseDSM] = useState(false)
  const [useVeg, setUseVeg] = useState(false)
  const [timezone, setTimezone] = useState('UTC+00:00')
  const [gridCells, setGridCells] = useState<GridCell[]>([])
  const [gridLoading, setGridLoading] = useState(false)
  const [comparePins, setComparePins] = useState<PinState[]>([])
  const [compareResults, setCompareResults] = useState<HorizonProfile[]>([])
  const [compareLoading, setCompareLoading] = useState(false)

  const handleRectangle = useCallback((bounds: { lat1: number; lng1: number; lat2: number; lng2: number }) => {
    setGridLoading(true)
    fetchGrid(bounds.lat1, bounds.lng1, bounds.lat2, bounds.lng2, height)
      .then(setGridCells)
      .catch(() => {})
      .finally(() => setGridLoading(false))
  }, [height])

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
      setTimezone(resolveTimezone(state.lat, state.lng))
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
    setTimezone(resolveTimezone(p.lat, p.lng))
  }, [])

  const workerPost = useWorker(
    () => new Worker(new URL('./lib/year.worker.ts', import.meta.url), { type: 'module' }),
    (workerData: WorkerYearResult) => {
      const days: DayResult[] = workerData.days.map((d) => ({
        date: new Date(d.date),
        dayOfYear: d.dayOfYear,
        totalMinutes: d.totalMinutes,
        firstSunMinute: d.firstSunMinute,
        lastSunMinute: d.lastSunMinute,
        sunStates: [],
      }))
      const yearResult: YearResult = { ...workerData, days }
      setYear(yearResult)
      setLoadState('loaded')
    },
    (_done: number, _total: number) => {}
  )

  useEffect(() => {
    if (!pin || offline) return
    setLoadState('loading')
    setLoadError('')

    Promise.all([
      fetchHorizon(pin.lat, pin.lng, height, useDSM, useVeg),
      fetchBuildings(pin.lat, pin.lng),
    ])
      .then(([p, b]) => {
        setProfile(p)
        setBuildings(b)
        workerPost({ type: 'compute', lat: pin.lat, lng: pin.lng, profile: p })
      })
      .catch((err) => {
        const msg = err.message || ''
        setLoadError(msg || 'Failed to compute. Please try again.')
        setLoadState('error')
      })
  }, [pin, height, offline, useDSM, useVeg])

  const handleAddToCompare = useCallback(() => {
    if (!pin) return
    setComparePins((prev) => {
      if (prev.find((p) => p.lat === pin.lat && p.lng === pin.lng)) return prev
      return [...prev, { lat: pin.lat, lng: pin.lng }]
    })
  }, [pin])

  const handleRemoveComparePin = useCallback((index: number) => {
    setComparePins((prev) => prev.filter((_, i) => i !== index))
    setCompareResults((prev) => prev.filter((_, i) => i !== index))
  }, [])

  const handleRunComparison = useCallback(async () => {
    if (comparePins.length < 2) return
    setCompareLoading(true)
    try {
      const results = await fetchBatchHorizon(
        comparePins.map((p) => ({ lat: p.lat, lng: p.lng, height, use_dsm: useDSM }))
      )
      setCompareResults(results.filter((r) => !r.error).map((r) => r.data!))
    } catch {
      setCompareResults([])
    } finally {
      setCompareLoading(false)
    }
  }, [comparePins, height, useDSM])

  const dayResult: DayResult | null = year && selectedDay >= 0 && selectedDay < year.days.length
    ? year.days[selectedDay]
    : null

  const sweepDate = pin
    ? new Date(Date.UTC(2025, 0, selectedDay + 1, Math.floor(timeFrac * 24), Math.floor((timeFrac * 24 - Math.floor(timeFrac * 24)) * 60)))
    : new Date()

  const summary = year && pin
    ? generateSummary(year, pin.lat, pin.lng, height, useDSM, timezone)
    : ''

  if (embed) {
    return (
      <div style={{ width: '100vw', height: '100vh', overflow: 'hidden', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
        <MapView
          pin={pin}
          onPinChange={handlePinChange}
          sweepDate={sweepDate}
          buildings={buildings}
          gridCells={gridCells.length > 0 ? gridCells : undefined}
          gridLoading={gridLoading}
          onRectangle={handleRectangle}
        />
        {loadState === 'loaded' && summary && (
          <div style={{
            position: 'absolute', bottom: 20, left: 20, right: 20,
            background: 'rgba(255,255,255,0.95)', padding: 12, borderRadius: 8,
            boxShadow: '0 2px 12px rgba(0,0,0,0.15)', fontSize: 12, lineHeight: 1.4,
          }}>
            {summary}
          </div>
        )}
        <InstallPrompt />
      </div>
    )
  }

  return (
    <div style={{ display: 'flex', width: '100vw', height: '100vh', overflow: 'hidden', fontFamily: 'system-ui, -apple-system, sans-serif' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <MapView
          pin={pin}
          onPinChange={handlePinChange}
          sweepDate={sweepDate}
          buildings={buildings}
          gridCells={gridCells.length > 0 ? gridCells : undefined}
          gridLoading={gridLoading}
          onRectangle={handleRectangle}
        />
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

        {pin && (
          <ProjectsPanel
            lat={pin.lat}
            lng={pin.lng}
            height={height}
            useDSM={useDSM}
            onLoadProject={(lat, lng, h, dsm) => {
              setPin({ lat, lng })
              setHeight(h)
              setUseDSM(dsm)
              setProfile(null)
              setYear(null)
              setLoadState('idle')
            }}
          />
        )}

        {offline && (
          <div style={{ padding: '10px 14px', background: '#fff3e0', border: '1px solid #ffcc80', borderRadius: 6, marginBottom: 16, fontSize: 13, color: '#e65100' }}>
            You are offline. Sunpath needs a network connection to fetch building data.
          </div>
        )}

        <div style={{ marginBottom: 16, display: 'flex', flexDirection: 'column', gap: 8, fontSize: 13 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: 6, cursor: 'pointer', color: '#555' }}>
              <input type="checkbox" checked={useDSM} onChange={(e) => setUseDSM(e.target.checked)} />
              Include terrain shadows
            </label>
            {useDSM && (
              <span style={{ fontSize: 11, color: '#e67e22', fontStyle: 'italic' }}>
                Open elevation data
              </span>
            )}
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: 6, cursor: 'pointer', color: '#555' }}>
              <input type="checkbox" checked={useVeg} onChange={(e) => setUseVeg(e.target.checked)} />
              Include vegetation shadows
            </label>
            {useVeg && (
              <span style={{ fontSize: 11, color: '#e67e22', fontStyle: 'italic' }}>
                Tree data may be limited
              </span>
            )}
          </div>
        </div>

        {pin && loadState === 'loaded' && profile && (
          <div style={{ display: 'flex', gap: 8, marginBottom: 16 }}>
            <a
              href={`/api/export/csv?lat=${pin.lat}&lng=${pin.lng}&h=${height}`}
              download
              style={{
                flex: 1, padding: '8px 12px', fontSize: 13, textAlign: 'center',
                border: '1px solid #3498db', borderRadius: 4, background: '#f0f8ff',
                color: '#3498db', cursor: 'pointer', textDecoration: 'none',
              }}
            >
              Download CSV
            </a>
            <a
              href={`/api/export/pdf?lat=${pin.lat}&lng=${pin.lng}&h=${height}`}
              download
              style={{
                flex: 1, padding: '8px 12px', fontSize: 13, textAlign: 'center',
                border: '1px solid #e67e22', borderRadius: 4, background: '#fff8f0',
                color: '#e67e22', cursor: 'pointer', textDecoration: 'none',
              }}
            >
              Download Report
            </a>
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
              timezone={timezone}
            />

            <PinInspector
              dayResult={dayResult}
              selectedDay={selectedDay}
              onDayChange={setSelectedDay}
              days={year.days}
              timezone={timezone}
            />

            <TimeSlider selectedDay={selectedDay} onTimeChange={setTimeFrac} timezone={timezone} />

            {profile && pin && (
              <SunIndicator date={sweepDate} lat={pin.lat} lng={pin.lng} profile={profile} />
            )}

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
        {pin && (
          <ComparisonPanel
            pins={comparePins}
            results={compareResults}
            loading={compareLoading}
            height={height}
            onAdd={handleAddToCompare}
            onRemove={handleRemoveComparePin}
            onRun={handleRunComparison}
            onNavigate={(lat, lng) => {
              setPin({ lat, lng })
              setProfile(null)
              setYear(null)
              setLoadState('idle')
            }}
          />
        )}
        <EmbedPanel />
        <AboutPanel />
      </div>
      <InstallPrompt />
    </div>
  )
}

export default App
