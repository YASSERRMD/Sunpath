import { getSunPosition } from './sun'
import type { HorizonProfile } from './api'

export function isInDirectSun(
  date: Date,
  lat: number,
  lng: number,
  profile: HorizonProfile
): boolean {
  const { azimuth, elevation } = getSunPosition(date, lat, lng)

  if (elevation <= 0) return false

  const azIndex = Math.round(azimuth) % 360
  return elevation > profile.horizon[azIndex]
}

export interface DayResult {
  date: Date
  dayOfYear: number
  totalMinutes: number
  firstSunMinute: number
  lastSunMinute: number
  sunStates: { time: Date; inSun: boolean }[]
}

export interface YearResult {
  days: DayResult[]
  grid: number[][]
  maxSunMinutes: number
  minSunMinutes: number
  bestDay: number
  worstDay: number
}

export function computeDay(
  date: Date,
  lat: number,
  lng: number,
  profile: HorizonProfile
): DayResult {
  const start = new Date(date)
  start.setUTCHours(0, 0, 0, 0)

  const end = new Date(start)
  end.setUTCDate(end.getUTCDate() + 1)

  let totalMinutes = 0
  let firstSunMinute = -1
  let lastSunMinute = -1
  const sunStates: { time: Date; inSun: boolean }[] = []

  const cursor = new Date(start)
  let minOfDay = 0
  while (cursor < end) {
    const inSun = isInDirectSun(cursor, lat, lng, profile)
    sunStates.push({ time: new Date(cursor), inSun })
    if (inSun) {
      totalMinutes++
      if (firstSunMinute < 0) firstSunMinute = minOfDay
      lastSunMinute = minOfDay
    }
    cursor.setUTCMinutes(cursor.getUTCMinutes() + 1)
    minOfDay++
  }

  const dayOfYear = Math.floor(
    (start.getTime() - new Date(start.getUTCFullYear(), 0, 0).getTime()) / 86400000
  )

  return { date: start, dayOfYear, totalMinutes, firstSunMinute, lastSunMinute, sunStates }
}

export function computeYear(
  lat: number,
  lng: number,
  profile: HorizonProfile
): YearResult {
  const days: DayResult[] = []
  const grid: number[][] = []

  let maxSunMinutes = 0
  let minSunMinutes = Infinity
  let bestDay = 0
  let worstDay = 0

  for (let doy = 0; doy < 365; doy++) {
    const date = new Date(Date.UTC(2025, 0, doy + 1))
    const result = computeDay(date, lat, lng, profile)
    days.push(result)

    const hourRow: number[] = new Array(24).fill(0)
    for (const st of result.sunStates) {
      const h = st.time.getUTCHours()
      if (st.inSun) hourRow[h] = 1
    }
    grid.push(hourRow)

    if (result.totalMinutes > maxSunMinutes) {
      maxSunMinutes = result.totalMinutes
      bestDay = doy
    }
    if (result.totalMinutes < minSunMinutes) {
      minSunMinutes = result.totalMinutes
      worstDay = doy
    }
  }

  return { days, grid, maxSunMinutes, minSunMinutes, bestDay, worstDay }
}
