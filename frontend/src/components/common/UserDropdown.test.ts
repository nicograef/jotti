import { describe, expect, it } from 'vitest'

import { moduswechselEintrag } from './UserDropdown'

describe('moduswechselEintrag', () => {
  it('bietet auf der Tischauswahl den Wechsel zum Direktverkauf an', () => {
    expect(moduswechselEintrag('/service/tische')).toEqual({
      label: 'Zu Direktverkauf wechseln',
      ziel: '/service/direktverkauf',
    })
  })

  it('bietet auf dem Tischdetail den Wechsel zum Direktverkauf an', () => {
    expect(moduswechselEintrag('/service/tische/12')).toEqual({
      label: 'Zu Direktverkauf wechseln',
      ziel: '/service/direktverkauf',
    })
  })

  it('bietet im Direktverkauf den Wechsel zum Tischservice an', () => {
    expect(moduswechselEintrag('/service/direktverkauf')).toEqual({
      label: 'Zu Tischservice wechseln',
      ziel: '/service/tische',
    })
  })

  it('erscheint nicht außerhalb des Service-Bereichs', () => {
    expect(moduswechselEintrag('/admin/produkte')).toBeNull()
  })
})
