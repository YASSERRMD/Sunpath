export function generateSummary(
  year: { days: { totalMinutes: number; firstSunMinute: number; lastSunMinute: number }[]; bestDay: number; worstDay: number },
  lat: number,
  lng: number,
  observerHeight: number,
  useDSM?: boolean
): string {
  const best = year.days[year.bestDay]
  const worst = year.days[year.worstDay]
  const fullyShaded = year.days.filter(d => d.totalMinutes === 0).length

  const fmt = (m: number) => {
    const h = Math.floor(m / 60)
    const min = m % 60
    return `${h}h${min > 0 ? min + 'm' : ''}`
  }

  const fmtRange = (d: { totalMinutes: number; firstSunMinute: number; lastSunMinute: number }) => {
    if (d.totalMinutes === 0 || d.firstSunMinute < 0) return 'no direct sun'
    const first = `${pad(Math.floor(d.firstSunMinute / 60))}:${pad(d.firstSunMinute % 60)}`
    const last = `${pad(Math.floor(d.lastSunMinute / 60))}:${pad(d.lastSunMinute % 60)}`
    return `${first} to ${last}`
  }

  const note = useDSM ? ' (includes terrain shadows).' : '.'
  return [
    `This point at ${lat.toFixed(4)}°N, ${lng.toFixed(4)}°E at ${observerHeight.toFixed(1)}m above ground`,
    `gets direct sun roughly ${fmtRange(best)} on the sunniest day (${fmt(best.totalMinutes)} min)`,
    `and ${fmtRange(worst)} on the shadiest day (${fmt(worst.totalMinutes)} min)${note}`,
    `It is fully shaded all day on ${fullyShaded} day${fullyShaded !== 1 ? 's' : ''} of the year.`,
  ].join(' ')
}

function pad(n: number): string {
  return n < 10 ? '0' + n : String(n)
}
