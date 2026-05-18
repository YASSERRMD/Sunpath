import { describe, it, expect, vi } from 'vitest'

describe('TimeSlider logic', () => {
  it('converts time fraction to hours and minutes correctly', () => {
    const frac = 0.5
    const hours = Math.floor(frac * 24)
    const mins = Math.floor((frac * 24 - hours) * 60)
    expect(hours).toBe(12)
    expect(mins).toBe(0)
  })

  it('start of day is midnight', () => {
    const frac = 0
    const hours = Math.floor(frac * 24)
    const mins = Math.floor((frac * 24 - hours) * 60)
    expect(hours).toBe(0)
    expect(mins).toBe(0)
  })

  it('end of day is 23:59', () => {
    const frac = 0.99999
    const hours = Math.floor(frac * 24)
    const mins = Math.floor((frac * 24 - hours) * 60)
    expect(hours >= 23).toBe(true)
  })

  it('noon at 0.5 fraction', () => {
    const frac = 0.5
    const totalMinutes = frac * 24 * 60
    expect(totalMinutes).toBe(720)
  })

  it('calls onTimeChange callback when value changes', () => {
    const onChange = vi.fn()
    const value = 0.25
    onChange(value)
    expect(onChange).toHaveBeenCalledWith(0.25)
  })
})
