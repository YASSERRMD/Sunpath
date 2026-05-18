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
  sunStates: { time: Date; inSun: boolean }[]
}

export interface YearResult {
  days: DayResult[]
  grid: number[][] // day x hour: 1=in sun, 0=in shade
  maxSunMinutes: number
  minSunMinutes: number
  bestDay: number
  worstDay: number
}
