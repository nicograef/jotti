import { describe, expect, it } from 'vitest'

import { ServiceTableGuard } from './routes'

describe('ServiceTableGuard', () => {
  it('redirects to table selection for invalid tischId', () => {
    const result = ServiceTableGuard({
      params: { tischId: 'abc' },
    } as never)

    expect(result).toBeInstanceOf(Response)
    if (!(result instanceof Response)) {
      throw new Error('Expected redirect response for invalid tischId')
    }
    expect(result.status).toBe(302)
    expect(result.headers.get('Location')).toBe('/service/tische')
  })

  it('allows valid positive integer tischId', () => {
    const result = ServiceTableGuard({
      params: { tischId: '12' },
    } as never)

    expect(result).toBeUndefined()
  })
})
