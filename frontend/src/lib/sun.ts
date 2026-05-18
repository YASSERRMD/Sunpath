import SunCalc from 'suncalc'

export interface SunPosition {
  azimuth: number
  elevation: number
}

export function getSunPosition(date: Date, lat: number, lng: number): SunPosition {
  const pos = SunCalc.getPosition(date, lat, lng)

  let az = pos.azimuth * 180 / Math.PI
  az = (az + 180) % 360
  if (az < 0) az += 360

  const elevation = pos.altitude * 180 / Math.PI

  return { azimuth: az, elevation }
}

export function getSunTimes(date: Date, lat: number, lng: number) {
  const times = SunCalc.getTimes(date, lat, lng)
  return times
}
