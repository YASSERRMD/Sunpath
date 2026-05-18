import { useState, useEffect, useCallback } from 'react'
import { fetchProjects, createProject, deleteProject, type ProjectRecord } from '../lib/api'
import { getToken, getStoredUser, setStoredUser, setToken, clearToken, isAuthenticated } from '../lib/auth'

interface ProjectsPanelProps {
  lat: number
  lng: number
  height: number
  useDSM: boolean
  onLoadProject: (lat: number, lng: number, height: number, useDSM: boolean) => void
}

export default function ProjectsPanel({ lat, lng, height, useDSM, onLoadProject }: ProjectsPanelProps) {
  const [projects, setProjects] = useState<ProjectRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [showLogin, setShowLogin] = useState(false)
  const [email, setEmail] = useState('')
  const [saving, setSaving] = useState(false)
  const [projectName, setProjectName] = useState('')
  const [authenticated, setAuthenticated] = useState(isAuthenticated())
  const [user, setUser] = useState(getStoredUser())
  const [expanded, setExpanded] = useState(false)

  const loadProjects = useCallback(async () => {
    const token = getToken()
    if (!token) return
    setLoading(true)
    try {
      const list = await fetchProjects(token)
      setProjects(list)
    } catch {
      clearToken()
      setAuthenticated(false)
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    if (authenticated && expanded) {
      loadProjects()
    }
  }, [authenticated, expanded, loadProjects])

  async function handleLogin() {
    if (!email) return
    try {
      const res = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: `email=${encodeURIComponent(email)}`,
      })
      const body = await res.json()
      if (body.error) {
        alert(body.error)
        return
      }
      const code = body.data?.code
      if (!code) {
        alert('Failed to get login code')
        return
      }
      const cbRes = await fetch(`/api/auth/callback?code=${code}`)
      const cbBody = await cbRes.json()
      if (cbBody.error) {
        alert(cbBody.error)
        return
      }
      const token: string = cbBody.data?.token
      const u = cbBody.data?.user
      setToken(token)
      setStoredUser(u)
      setAuthenticated(true)
      setUser(u)
      setShowLogin(false)
      setEmail('')
    } catch (err) {
      alert('Login failed: ' + (err instanceof Error ? err.message : String(err)))
    }
  }

  async function handleSave() {
    const token = getToken()
    if (!token) return
    setSaving(true)
    try {
      const name = projectName || `Point (${lat.toFixed(4)}, ${lng.toFixed(4)})`
      await createProject(token, { name, lat, lng, height, use_dsm: useDSM })
      setProjectName('')
      await loadProjects()
    } catch (err) {
      alert('Failed to save: ' + (err instanceof Error ? err.message : String(err)))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(id: number) {
    const token = getToken()
    if (!token) return
    try {
      await deleteProject(token, id)
      setProjects(projects.filter((p) => p.id !== id))
    } catch {
      alert('Failed to delete project')
    }
  }

  function handleLogout() {
    clearToken()
    setAuthenticated(false)
    setUser(null)
    setProjects([])
  }

  return (
    <div style={{ marginBottom: 16 }}>
      <button
        onClick={() => setExpanded(!expanded)}
        style={{
          width: '100%', padding: '8px 12px', fontSize: 13, fontWeight: 500,
          border: '1px solid #ccc', borderRadius: 4, background: '#f8f8f8',
          cursor: 'pointer', textAlign: 'left', display: 'flex', justifyContent: 'space-between',
        }}
      >
        <span>{authenticated ? (user ? user.email : 'Projects') : 'Sign in to save projects'}</span>
        <span>{expanded ? '▲' : '▼'}</span>
      </button>

      {expanded && (
        <div style={{ marginTop: 8, fontSize: 13 }}>
          {!authenticated && !showLogin && (
            <button
              onClick={() => setShowLogin(true)}
              style={{ width: '100%', padding: '6px 12px', border: '1px solid #3498db', borderRadius: 4, background: '#3498db', color: '#fff', cursor: 'pointer', fontSize: 13 }}
            >
              Sign in with email
            </button>
          )}

          {!authenticated && showLogin && (
            <div style={{ display: 'flex', gap: 6 }}>
              <input
                type="email" placeholder="your@email.com" value={email}
                onChange={(e) => setEmail(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleLogin()}
                style={{ flex: 1, padding: '6px 8px', border: '1px solid #ccc', borderRadius: 4, fontSize: 13 }}
              />
              <button
                onClick={handleLogin}
                style={{ padding: '6px 12px', border: '1px solid #27ae60', borderRadius: 4, background: '#27ae60', color: '#fff', cursor: 'pointer', fontSize: 13 }}
              >
                Go
              </button>
            </div>
          )}

          {authenticated && user && (
            <div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8, color: '#666' }}>
                <span>{user.email}</span>
                <button onClick={handleLogout} style={{ border: 'none', background: 'none', color: '#e74c3c', cursor: 'pointer', fontSize: 12, textDecoration: 'underline' }}>
                  Sign out
                </button>
              </div>

              <div style={{ display: 'flex', gap: 6, marginBottom: 8 }}>
                <input
                  type="text" placeholder="Project name" value={projectName}
                  onChange={(e) => setProjectName(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleSave()}
                  style={{ flex: 1, padding: '6px 8px', border: '1px solid #ccc', borderRadius: 4, fontSize: 13 }}
                />
                <button
                  onClick={handleSave} disabled={saving}
                  style={{
                    padding: '6px 12px', border: '1px solid #27ae60', borderRadius: 4, background: '#27ae60', color: '#fff', cursor: 'pointer', fontSize: 13,
                    opacity: saving ? 0.6 : 1,
                  }}
                >
                  {saving ? '...' : 'Save'}
                </button>
              </div>

              {loading && <div style={{ color: '#999', padding: '4px 0' }}>Loading...</div>}

              {!loading && projects.length === 0 && (
                <div style={{ color: '#999', padding: '4px 0' }}>No saved projects yet.</div>
              )}

              {projects.map((p) => (
                <div key={p.id} style={{
                  display: 'flex', justifyContent: 'space-between', alignItems: 'center',
                  padding: '6px 8px', borderBottom: '1px solid #eee', cursor: 'pointer',
                }}>
                  <div
                    onClick={() => onLoadProject(p.lat, p.lng, p.height, p.use_dsm)}
                    style={{ flex: 1, overflow: 'hidden' }}
                  >
                    <div style={{ fontWeight: 500, fontSize: 13 }}>{p.name}</div>
                    <div style={{ fontSize: 11, color: '#999' }}>
                      {p.lat.toFixed(4)}, {p.lng.toFixed(4)} | h={p.height}m{p.use_dsm ? ' +DSM' : ''}
                    </div>
                  </div>
                  <button
                    onClick={(e) => { e.stopPropagation(); handleDelete(p.id) }}
                    style={{ border: 'none', background: 'none', color: '#e74c3c', cursor: 'pointer', fontSize: 14, padding: '2px 4px' }}
                    title="Delete"
                  >
                    x
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
