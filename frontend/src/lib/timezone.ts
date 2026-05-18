export function resolveTimezone(_lat: number, lng: number): string {
  const offsetHours = Math.round(lng / 15)
  const sign = offsetHours >= 0 ? '+' : '-'
  const abs = Math.abs(offsetHours)
  const padded = String(abs).padStart(2, '0')
  return `Etc/GMT${sign}${padded}`
}

export function formatLocalTime(date: Date, timezone: string): string {
  try {
    return date.toLocaleTimeString('en-US', {
      timeZone: timezone,
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
    })
  } catch {
    return date.toISOString().slice(11, 16)
  }
}

export function applyTimezone(date: Date, timezone: string): Date {
  const localStr = date.toLocaleString('en-US', { timeZone: timezone })
  return new Date(localStr)
}
