interface StorageValue<T> {
  value: T
  expiresAt?: number
}

export const storage = {
  get<T>(key: string): T | null {
    const raw = window.localStorage.getItem(key)
    if (!raw) return null

    try {
      const parsed = JSON.parse(raw) as StorageValue<T>
      if (parsed.expiresAt && Date.now() > parsed.expiresAt) {
        window.localStorage.removeItem(key)
        return null
      }
      return parsed.value
    } catch {
      window.localStorage.removeItem(key)
      return null
    }
  },
  set<T>(key: string, value: T, ttl?: number): void {
    const payload: StorageValue<T> = {
      value,
      expiresAt: ttl ? Date.now() + ttl : undefined,
    }
    window.localStorage.setItem(key, JSON.stringify(payload))
  },
  remove(key: string): void {
    window.localStorage.removeItem(key)
  },
}
