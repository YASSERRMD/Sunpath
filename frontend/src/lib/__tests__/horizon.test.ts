import { describe, it, expect } from 'vitest'
import { isInDirectSun, computeDay, computeYear } from '../horizon'
import type { HorizonProfile } from '../api'

function makeOpenFieldProfile(): HorizonProfile {
  return {
    horizon: new Array(360).fill(0),
    lat: 48.8566,
    lng: 2.3522,
    observer_height: 1.5,
    confidence: 1,
    building_count: 0,
    estimated_count: 0,
    data_hash: 'open',
  }
}

function makeFullyObstructedProfile(): HorizonProfile {
  const horizon = new Array(360).fill(90)
  return {
    horizon,
    lat: 48.8566,
    lng: 2.3522,
    observer_height: 1.5,
    confidence: 0.5,
    building_count: 10,
    estimated_count: 5,
    data_hash: 'boxed',
  }
}

describe('isInDirectSun', () => {
  it('returns true for daytime in open field', () => {
    const profile = makeOpenFieldProfile()
    const date = new Date(Date.UTC(2025, 5, 20, 12, 0, 0))
    expect(isInDirectSun(date, 48.8566, 2.3522, profile)).toBe(true)
  })

  it('returns false at night', () => {
    const profile = makeOpenFieldProfile()
    const date = new Date(Date.UTC(2025, 5, 20, 0, 0, 0))
    expect(isInDirectSun(date, 48.8566, 2.3522, profile)).toBe(false)
  })

  it('returns false when fully obstructed', () => {
    const profile = makeFullyObstructedProfile()
    const date = new Date(Date.UTC(2025, 5, 20, 12, 0, 0))
    expect(isInDirectSun(date, 48.8566, 2.3522, profile)).toBe(false)
  })
})

describe('computeDay', () => {
  it('returns correct sun hours for open field at summer solstice', () => {
    const profile = makeOpenFieldProfile()
    const date = new Date(Date.UTC(2025, 5, 20))
    const result = computeDay(date, 48.8566, 2.3522, profile)

    expect(result.totalMinutes).toBeGreaterThan(600)
    expect(result.totalMinutes).toBeLessThan(960)
    expect(result.sunStates.length).toBe(1440)
  })

  it('returns 0 sun minutes for fully obstructed', () => {
    const profile = makeFullyObstructedProfile()
    const date = new Date(Date.UTC(2025, 5, 20))
    const result = computeDay(date, 48.8566, 2.3522, profile)

    expect(result.totalMinutes).toBe(0)
  })

  it('has correct day of year', () => {
    const profile = makeOpenFieldProfile()
    const date = new Date(Date.UTC(2025, 0, 1))
    const result = computeDay(date, 48.8566, 2.3522, profile)

    expect(result.dayOfYear).toBe(1)
  })
})

describe('computeYear', () => {
  it('returns 365 days', () => {
    const profile = makeOpenFieldProfile()
    const result = computeYear(48.8566, 2.3522, profile)

    expect(result.days.length).toBe(365)
    expect(result.grid.length).toBe(365)
  })

  it('best day has more sun than worst day', () => {
    const profile = makeOpenFieldProfile()
    const result = computeYear(48.8566, 2.3522, profile)

    expect(result.maxSunMinutes).toBeGreaterThan(result.minSunMinutes)
  })

  it('grid cells are 0 or 1', () => {
    const profile = makeOpenFieldProfile()
    const result = computeYear(48.8566, 2.3522, profile)

    for (const row of result.grid) {
      for (const cell of row) {
        expect(cell === 0 || cell === 1).toBe(true)
      }
    }
  })
})
