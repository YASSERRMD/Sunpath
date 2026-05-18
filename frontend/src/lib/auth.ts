const TOKEN_KEY = 'sunpath_auth_token'
const USER_KEY = 'sunpath_auth_user'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

export function isAuthenticated(): boolean {
  return getToken() !== null
}

export function getStoredUser(): { id: number; email: string; name: string } | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw)
  } catch {
    return null
  }
}

export function setStoredUser(user: { id: number; email: string; name: string }): void {
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}
