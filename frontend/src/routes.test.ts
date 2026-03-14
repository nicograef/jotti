import { describe, expect, it } from 'vitest'

import { ServiceTableGuard } from './routes'

describe('ServiceTableGuard', () => {
  it('redirects to table selection for invalid tableId', () => {
    const result = ServiceTableGuard({
      params: { tableId: 'abc' },
    } as never)

    expect(result).toBeInstanceOf(Response)
    if (!(result instanceof Response)) {
      throw new Error('Expected redirect response for invalid tableId')
    }
    expect(result.status).toBe(302)
    expect(result.headers.get('Location')).toBe('/service/tables')
  })

  it('allows valid positive integer tableId', () => {
    const result = ServiceTableGuard({
      params: { tableId: '12' },
    } as never)

    expect(result).toBeUndefined()
  })
})
