import React, { useRef, useEffect, useMemo } from 'react'
import { formatMinutes } from '../lib/timezone'

interface YearHeatmapProps {
  grid: number[][]
  selectedDay: number
  onDaySelect: (day: number) => void
  timezone?: string
}

const CELL_W = 2
const CELL_H = 12
const GAP = 0.5
const MARGIN = { top: 20, right: 10, bottom: 30, left: 40 }

export default function YearHeatmap({ grid, selectedDay, onDaySelect, timezone }: YearHeatmapProps) {
  const tz = timezone || 'UTC+00:00'
  const canvasRef = useRef<HTMLCanvasElement>(null)

  const dims = useMemo(() => {
    const w = MARGIN.left + MARGIN.right + 365 * (CELL_W + GAP)
    const h = MARGIN.top + MARGIN.bottom + 24 * (CELL_H + GAP)
    return { w, h }
  }, [])

  const monthLabels = useMemo(() => {
    const labels: { x: number; label: string }[] = []
    for (let m = 0; m < 12; m++) {
      const doy = Math.floor((new Date(2025, m, 1).getTime() - new Date(2025, 0, 1).getTime()) / 86400000)
      const x = MARGIN.left + doy * (CELL_W + GAP)
      labels.push({ x, label: ['J','F','M','A','M','J','J','A','S','O','N','D'][m] })
    }
    return labels
  }, [])

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || grid.length === 0) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    canvas.width = dims.w
    canvas.height = dims.h

    ctx.clearRect(0, 0, dims.w, dims.h)

    for (let doy = 0; doy < 365 && doy < grid.length; doy++) {
      for (let hour = 0; hour < 24; hour++) {
        const x = MARGIN.left + doy * (CELL_W + GAP)
        const y = MARGIN.top + hour * (CELL_H + GAP)
        const isSun = grid[doy]?.[hour] === 1

        ctx.fillStyle = isSun
          ? 'hsl(35, 90%, 55%)'
          : 'hsl(220, 15%, 85%)'

        ctx.fillRect(x, y, CELL_W, CELL_H)
      }
    }

    ctx.fillStyle = '#e74c3c'
    const sx = MARGIN.left + selectedDay * (CELL_W + GAP)
    ctx.fillRect(sx - 1, MARGIN.top - 2, CELL_W + 2, 24 * (CELL_H + GAP) + 4)

    ctx.fillStyle = '#333'
    ctx.font = '10px system-ui, sans-serif'
    ctx.textAlign = 'center'
    for (const ml of monthLabels) {
      ctx.fillText(ml.label, ml.x + CELL_W / 2, MARGIN.top - 4)
    }

    ctx.textAlign = 'right'
    for (let h = 0; h < 24; h += 3) {
      const y = MARGIN.top + h * (CELL_H + GAP) + CELL_H / 2 + 3
      const label = formatMinutes(h * 60, tz)
      ctx.fillText(label, MARGIN.left - 4, y)
    }

    ctx.fillStyle = '#999'
    ctx.textAlign = 'center'
    ctx.fillText('Day of year', dims.w / 2, dims.h - 6)
  }, [grid, selectedDay, dims, monthLabels, tz])

  const handleClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const rect = canvasRef.current?.getBoundingClientRect()
    if (!rect) return
    const x = e.clientX - rect.left - MARGIN.left
    const doy = Math.round(x / (CELL_W + GAP))
    if (doy >= 0 && doy < 365) {
      onDaySelect(doy)
    }
  }

  return (
    <canvas
      ref={canvasRef}
      onClick={handleClick}
      style={{ width: '100%', maxWidth: dims.w, cursor: 'pointer', marginBottom: 16 }}
    />
  )
}
