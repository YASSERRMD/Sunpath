import { describe, it, expect } from 'vitest'

describe('Grid logic', () => {
  it('finds best cell correctly', () => {
    const cells = [
      { lat: 0, lng: 0, sun_minutes: 300 },
      { lat: 1, lng: 1, sun_minutes: 500 },
      { lat: 2, lng: 2, sun_minutes: 100 },
    ]
    const best = cells.reduce((a, b) => a.sun_minutes > b.sun_minutes ? a : b)
    expect(best.sun_minutes).toBe(500)
    expect(best.lat).toBe(1)
  })

  it('finds worst cell correctly', () => {
    const cells = [
      { lat: 0, lng: 0, sun_minutes: 300 },
      { lat: 1, lng: 1, sun_minutes: 500 },
      { lat: 2, lng: 2, sun_minutes: 100 },
    ]
    const worst = cells.reduce((a, b) => a.sun_minutes < b.sun_minutes ? a : b)
    expect(worst.sun_minutes).toBe(100)
    expect(worst.lat).toBe(2)
  })

  it('handles single cell', () => {
    const cells = [{ lat: 0, lng: 0, sun_minutes: 400 }]
    const best = cells.reduce((a, b) => a.sun_minutes > b.sun_minutes ? a : b)
    expect(best.sun_minutes).toBe(400)
  })

  it('intensity is between 0 and 1', () => {
    const cells = [
      { lat: 0, lng: 0, sun_minutes: 100 },
      { lat: 1, lng: 1, sun_minutes: 800 },
    ]
    const maxMin = Math.max(...cells.map(c => c.sun_minutes), 1)
    const intensities = cells.map(c => c.sun_minutes / maxMin)
    expect(intensities[0]).toBeCloseTo(0.125)
    expect(intensities[1]).toBeCloseTo(1)
  })
})
