import type { YearResult, DayResult } from './horizon'

export function generateSummary(
  year: YearResult,
  lat: number,
  lng: number,
  observerHeight: number
): string {
  const best = year.days[year.bestDay]
  const worst = year.days[year.worstDay]
  const fullyShaded = year.days.filter(d => d.totalMinutes === 0).length

  const fmt = (d: DayResult) => {
    const h = Math.floor(d.totalMinutes / 60)
    const m = d.totalMinutes % 60
    return `${h}h${m > 0 ? m + 'm' : ''}`
  }

  const fmtRange = (d: DayResult) => {
    const first = d.sunStates.find(s => s.inSun)
    const last = [...d.sunStates].reverse().find(s => s.inSun)
    if (!first || !last) return 'no direct sun'
    return `${first.time.toISOString().slice(11, 16)} to ${last.time.toISOString().slice(11, 16)}`
  }

  return [
    `This point at ${lat.toFixed(4)}°N, ${lng.toFixed(4)}°E at ${observerHeight.toFixed(1)}m above ground`,
    `gets direct sun roughly ${fmtRange(best)} on the sunniest day (${fmt(best)})`,
    `and ${fmtRange(worst)} on the shadiest day (${fmt(worst)}).`,
    `It is fully shaded all day on ${fullyShaded} day${fullyShaded !== 1 ? 's' : ''} of the year.`,
  ].join(' ')
}
