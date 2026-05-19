export function resolveTimezone(_lat: number, lng: number): string {
  const offsetHours = Math.round(lng / 15)
  const sign = offsetHours >= 0 ? '+' : '-'
  const abs = Math.abs(offsetHours)
  const padded = String(abs).padStart(2, '0')
  return `UTC${sign}${padded}:00`
}

export function formatLocalTime(date: Date, timezone: string): string {
  const offsetMinutes = parseTimezoneOffset(timezone)
  const local = new Date(date.getTime() + offsetMinutes * 60000)
  const h = String(local.getUTCHours()).padStart(2, '0')
  const m = String(local.getUTCMinutes()).padStart(2, '0')
  return `${h}:${m}`
}

export function applyTimezone(date: Date, timezone: string): Date {
  const offsetMinutes = parseTimezoneOffset(timezone)
  return new Date(date.getTime() + offsetMinutes * 60000)
}

export function formatMinutes(minutes: number, timezone: string): string {
  const offsetMinutes = parseTimezoneOffset(timezone)
  const localMinutes = ((minutes + offsetMinutes) % 1440 + 1440) % 1440
  const h = String(Math.floor(localMinutes / 60)).padStart(2, '0')
  const m = String(Math.floor(localMinutes % 60)).padStart(2, '0')
  return `${h}:${m}`
}

function parseTimezoneOffset(tz: string): number {
  const match = tz.match(/^UTC([+-])(\d{2}):(\d{2})$/)
  if (!match) return 0
  const sign = match[1] === '+' ? 1 : -1
  const h = parseInt(match[2], 10)
  const m = parseInt(match[3], 10)
  return sign * (h * 60 + m)
}
