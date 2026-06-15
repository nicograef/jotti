const STORAGE_KEY = 'jotti-arbeitsmodus'

export type Arbeitsmodus = 'tischservice' | 'direktverkauf'

const DEFAULT_ARBEITSMODUS: Arbeitsmodus = 'tischservice'

// Geräte-Präferenz: überlebt Logout/Login und ist pro Gerät (BYOD).
export function getArbeitsmodus(): Arbeitsmodus {
  const stored = localStorage.getItem(STORAGE_KEY)
  return stored === 'tischservice' || stored === 'direktverkauf'
    ? stored
    : DEFAULT_ARBEITSMODUS
}

export function setArbeitsmodus(modus: Arbeitsmodus): void {
  localStorage.setItem(STORAGE_KEY, modus)
}
