interface ConfidenceBannerProps {
  confidence: number
  buildingCount: number
  estimatedCount: number
}

export default function ConfidenceBanner({ confidence, buildingCount, estimatedCount }: ConfidenceBannerProps) {
  if (confidence >= 0.7) return null

  const pct = Math.round((1 - confidence) * 100)

  return (
    <div style={{
      padding: '10px 14px',
      background: '#fff8e1',
      border: '1px solid #ffe082',
      borderRadius: 6,
      marginBottom: 16,
      fontSize: 13,
      color: '#8d6e00',
      lineHeight: 1.4,
    }}>
      <strong>Low confidence result</strong>
      <br />
      {pct}% of nearby buildings ({estimatedCount} of {buildingCount}) have estimated heights
      because OpenStreetMap data is thin in this area. Sun/shade boundaries may be
      less accurate than in areas with complete building data.
    </div>
  )
}
