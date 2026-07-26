import { cleanup, render } from '@testing-library/react'
import { useForm } from 'react-hook-form'
import { afterEach, beforeAll, describe, expect, it } from 'vitest'

import { drainInputOtpTimers } from '@/test/input-otp'

import { OTPField } from './FormFields'

// input-otp registriert intern einen ResizeObserver; jsdom bringt keinen mit.
class ResizeObserverStub {
  observe() {
    /* no-op */
  }
  unobserve() {
    /* no-op */
  }
  disconnect() {
    /* no-op */
  }
}

beforeAll(() => {
  globalThis.ResizeObserver = ResizeObserverStub
})

afterEach(async () => {
  cleanup()
  await drainInputOtpTimers()
})

function Harness() {
  const form = useForm<{ onetimePassword: string }>({
    defaultValues: { onetimePassword: '' },
  })
  return <OTPField form={form} />
}

describe('OTPField', () => {
  it('zeigt 6 Ziffern-Slots mit numerischer Eingabe', () => {
    const { container } = render(<Harness />)

    const slots = container.querySelectorAll('[data-slot="input-otp-slot"]')
    expect(slots).toHaveLength(6)

    const input = container.querySelector('[data-input-otp]')
    expect(input).not.toBeNull()
    expect(input).toHaveAttribute('maxlength', '6')
    expect(input).toHaveAttribute('inputmode', 'numeric')
    expect(input).toHaveAttribute('autocomplete', 'one-time-code')
    expect(input).toHaveAttribute('pattern', '^\\d+$')
  })
})
