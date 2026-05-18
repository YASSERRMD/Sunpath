import { useState } from 'react'

export default function AboutPanel() {
  const [open, setOpen] = useState(false)

  return (
    <div style={{ marginTop: 24, borderTop: '1px solid #e0e0e0', paddingTop: 16 }}>
      <button
        onClick={() => setOpen(!open)}
        style={{
          background: 'none',
          border: 'none',
          color: '#999',
          fontSize: 12,
          cursor: 'pointer',
          textDecoration: 'underline',
          padding: 0,
        }}
      >
        {open ? 'Hide' : 'About'} the method and its limits
      </button>

      {open && (
        <div style={{ fontSize: 12, color: '#777', lineHeight: 1.6, marginTop: 8 }}>
          <p><strong>How it works</strong></p>
          <p>
            Sunpath uses a 2.5D shadow model. Buildings from OpenStreetMap are treated as
            extruded prisms with flat roofs. For each compass direction (0-359 degrees), the
            engine computes the highest elevation angle at which a building edge blocks the sky.
            A point is in direct sun when the sun's elevation exceeds this horizon angle.
          </p>
          <p><strong>Building heights</strong></p>
          <p>
            OSM building height data is inconsistent. Heights are resolved in this priority:
            explicit height tag, building:levels tag (assumed 3.2m per level), or a default of 8m.
            Estimated heights are flagged and reduce the confidence score shown on the results.
          </p>
          <p><strong>Limitations</strong></p>
          <p>
            This is a 2.5D model. It does not account for terrain elevation, trees, or other
            vegetation. It works best in urban areas with good OSM building coverage. In areas
            with thin data, results are labelled low confidence. The analysis is deterministic
            and uses standard solar astronomy. No AI or machine learning is involved.
          </p>
          <p><strong>v1.0 non-goals</strong></p>
          <p>
            No user accounts, no saved projects (state is shared via URL only), no mobile native
            apps, no global coverage guarantee. Performance depends on OSM data quality in the
            region.
          </p>
        </div>
      )}
    </div>
  )
}
