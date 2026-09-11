const MARKET_BASE_URL_KEY = 'novablog_market_base_url'

// createStorage 创建基于指定后端存储与 key 前缀的存取集合。
function createStorage(backend: Storage, prefix: string) {
  const TOKEN_KEY = `${prefix}token`
  const REFRESH_TOKEN_KEY = `${prefix}refresh_token`
  const USER_KEY = `${prefix}user`
  const INITIALIZED_KEY = `${prefix}initialized`

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

export const storage = createStorage(localStorage, 'novablog_')
export const storageSession = createStorage(sessionStorage, 'novablog_')

const marketBase = createStorage(localStorage, 'novablog_market_')

// marketStorage 官方主题市场账号凭据存储（与后台管理员凭据相互隔离）。
export const marketStorage = {
  ...marketBase,
  getBaseURL(): string | null { return localStorage.getItem(MARKET_BASE_URL_KEY) },
  setBaseURL(url: string) { localStorage.setItem(MARKET_BASE_URL_KEY, url) },
  removeBaseURL() { localStorage.removeItem(MARKET_BASE_URL_KEY) },
  clear() {
    marketBase.clear()
    localStorage.removeItem(MARKET_BASE_URL_KEY)
  },
}
