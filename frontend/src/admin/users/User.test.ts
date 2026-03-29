import { describe, expect, it } from 'vitest'

import { UserSchema } from './User'

const validUser = {
  id: 1,
  name: 'Test User',
  username: 'testuser',
  role: 'service' as const,
  status: 'active' as const,
  createdAt: '2024-01-01T00:00:00Z',
  updatedAt: '2024-01-02T00:00:00Z',
}

describe('UserSchema', () => {
  it('accepts active users', () => {
    expect(UserSchema.safeParse(validUser).success).toBe(true)
  })

  it('accepts inactive users', () => {
    const result = UserSchema.safeParse({ ...validUser, status: 'inactive' })
    expect(result.success).toBe(true)
  })

  it('rejects deleted status — backend query filters deleted users before response', () => {
    const result = UserSchema.safeParse({ ...validUser, status: 'deleted' })
    expect(result.success).toBe(false)
  })

  it('requires updatedAt', () => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const { updatedAt: _, ...withoutUpdatedAt } = validUser
    const result = UserSchema.safeParse(withoutUpdatedAt)
    expect(result.success).toBe(false)
  })
})
