import { useState, useCallback } from 'react'
import MapView from './components/MapView'
import SidePanel from './components/SidePanel'

export interface PinState {
  lat: number
  lng: number
}

function App() {
  const [pin, setPin] = useState<PinState | null>(null)
  const [height, setHeight] = useState(1.5)

  const handlePinChange = useCallback((p: PinState) => {
    setPin(p)
  }, [])

  return (
    <div style={{ display: 'flex', width: '100vw', height: '100vh', overflow: 'hidden' }}>
      <div style={{ flex: 1, position: 'relative' }}>
        <MapView pin={pin} onPinChange={handlePinChange} />
      </div>
      <SidePanel pin={pin} height={height} onHeightChange={setHeight} onPinChange={handlePinChange} />
    </div>
  )
}

export default App
