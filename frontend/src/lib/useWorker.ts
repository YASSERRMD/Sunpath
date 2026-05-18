import { useEffect, useRef, useCallback } from 'react'

type WorkerResult<T> = { type: 'result'; data: T } | { type: 'progress'; done: number; total: number }

export function useWorker<T>(
  workerFactory: () => Worker,
  onResult: (data: T) => void,
  onProgress?: (done: number, total: number) => void
) {
  const workerRef = useRef<Worker | null>(null)
  const factoryRef = useRef(workerFactory)
  const onResultRef = useRef(onResult)
  const onProgressRef = useRef(onProgress)

  factoryRef.current = workerFactory
  onResultRef.current = onResult
  onProgressRef.current = onProgress

  useEffect(() => {
    const w = factoryRef.current()
    workerRef.current = w

    w.onmessage = (e: MessageEvent<WorkerResult<T>>) => {
      if (e.data.type === 'result') {
        onResultRef.current(e.data.data)
      } else if (e.data.type === 'progress') {
        onProgressRef.current?.(e.data.done, e.data.total)
      }
    }

    return () => {
      w.terminate()
      workerRef.current = null
    }
  }, [])

  const post = useCallback((msg: unknown) => {
    if (workerRef.current) {
      workerRef.current.postMessage(msg)
    }
  }, [])

  return post
}
