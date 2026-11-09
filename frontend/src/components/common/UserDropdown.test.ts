import { describe, expect, it } from 'vitest'

import { moduswechselEintrag } from './UserDropdown'

const zuDirektverkauf = {
  label: 'Zu Direktverkauf wechseln',
  ziel: '/service/direktverkauf',
}
const zuTischservice = {
  label: 'Zu Tischservice wechseln',
  ziel: '/service/tische',
}

describe('moduswechselEintrag', () => {
  it.each([
    ['/service/tische', zuDirektverkauf],
    ['/service/tische/12', zuDirektverkauf],
    ['/service/direktverkauf', zuTischservice],
    ['/admin/produkte', null],
  ])('bietet auf %s den passenden Moduswechsel an', (pfad, eintrag) => {
    expect(moduswechselEintrag(pfad)).toEqual(eintrag)
  })
})
