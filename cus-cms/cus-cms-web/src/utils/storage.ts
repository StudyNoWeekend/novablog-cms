const TOKEN_KEY = 'cus_cms_token'
const REFRESH_TOKEN_KEY = 'cus_cms_refresh_token'
const USER_KEY = 'cus_cms_user'
const INITIALIZED_KEY = 'cus_cms_initialized'

function createStorage(backend: Storage) {
  return {
    getToken(): string | null { return backend.getItem(TOKEN_KEY) },
    setToken(token: string) { backend.setItem(TOKEN_KEY, token) },
    removeToken() { backend.removeItem(TOKEN_KEY) },
    getRefreshToken(): string | null { return backend.getItem(REFRESH_TOKEN_KEY) },
    setRefreshToken(token: string) { backend.setItem(REFRESH_TOKEN_KEY, token) },
    removeRefreshToken() { backend.removeItem(REFRESH_TOKEN_KEY) },
    getUser(): any { const v = backend.getItem(USER_KEY); return v ? JSON.parse(v) : null },
    setUser(user: any) { backend.setItem(USER_KEY, JSON.stringify(user)) },
    removeUser() { backend.removeItem(USER_KEY) },
    getInitialized(): boolean | null {
      const v = backend.getItem(INITIALIZED_KEY)
      if (v === null) return null
      return v === 'true'
    },
    setInitialized(value: boolean) { backend.setItem(INITIALIZED_KEY, String(value)) },
    removeInitialized() { backend.removeItem(INITIALIZED_KEY) },
    clear() {
      backend.removeItem(TOKEN_KEY)
      backend.removeItem(REFRESH_TOKEN_KEY)
      backend.removeItem(USER_KEY)
      backend.removeItem(INITIALIZED_KEY)
    },
  }
}

export const storage = createStorage(localStorage)
export const storageSession = createStorage(sessionStorage)