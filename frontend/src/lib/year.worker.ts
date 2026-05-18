import { getSunPosition } from './sun'

interface WorkerProfile {
  horizon: number[]
  lat: number
  lng: number
  observer_height: number
  confidence: number
  building_count: number
  estimated_count: number
  data_hash: string
}

interface WorkerInput {
  type: 'compute'
  lat: number
  lng: number
  profile: WorkerProfile
}

interface DayResult {
  date: string
  dayOfYear: number
  totalMinutes: number
  firstSunMinute: number
  lastSunMinute: number
}

interface YearResult {
  days: DayResult[]
  grid: number[][]
  maxSunMinutes: number
  minSunMinutes: number
  bestDay: number
  worstDay: number
}

self.onmessage = (e: MessageEvent<WorkerInput>) => {
  if (e.data.type !== 'compute') return

  const { lat, lng, profile } = e.data
  const horizon = profile.horizon

  function isInDirectSun(date: Date): boolean {
    const pos = getSunPosition(date, lat, lng)
    if (pos.elevation <= 0) return false
    const azIndex = Math.round(pos.azimuth) % 360
    return pos.elevation > horizon[azIndex]
  }

  const days: DayResult[] = []
  const grid: number[][] = []
  let maxSunMinutes = 0
  let minSunMinutes = Infinity
  let bestDay = 0
  let worstDay = 0

  for (let doy = 0; doy < 365; doy++) {
    const date = new Date(Date.UTC(2025, 0, doy + 1))
    let totalMinutes = 0
    let firstSunMinute = -1
    let lastSunMinute = -1
    const hourRow: number[] = new Array(24).fill(0)

    for (let min = 0; min < 1440; min++) {
      const t = new Date(date)
      t.setUTCMinutes(min)
      const inSun = isInDirectSun(t)
      if (inSun) {
        totalMinutes++
        if (firstSunMinute < 0) firstSunMinute = min
        lastSunMinute = min
        const h = Math.floor(min / 60)
        hourRow[h] = 1
      }
    }

    days.push({ date: date.toISOString(), dayOfYear: doy, totalMinutes, firstSunMinute, lastSunMinute })
    grid.push(hourRow)

    if (totalMinutes > maxSunMinutes) {
      maxSunMinutes = totalMinutes
      bestDay = doy
    }
    if (totalMinutes < minSunMinutes) {
      minSunMinutes = totalMinutes
      worstDay = doy
    }

    if (doy % 30 === 0) {
      self.postMessage({ type: 'progress', done: doy + 1, total: 365 })
    }
  }

  const result: YearResult = { days, grid, maxSunMinutes, minSunMinutes, bestDay, worstDay }
  self.postMessage({ type: 'result', data: result })
}
