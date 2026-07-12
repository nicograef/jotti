import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { KommentarField } from './CommentField'

afterEach(() => {
  cleanup()
})

// Spiegelt die Verdrahtung der Storno-Drawer: invalid, solange getrimmt < 3 Zeichen.
function RequiredHarness() {
  const [kommentar, setKommentar] = useState('')
  return (
    <KommentarField
      required
      invalid={kommentar.trim().length < 3}
      onChange={setKommentar}
    />
  )
}

const HINWEIS = 'Kommentar ist erforderlich (mind. 3 Zeichen).'

describe('KommentarField', () => {
  it('zeigt den Pflichthinweis bei required von Anfang an muted, ohne Berührung', () => {
    render(<RequiredHarness />)

    const hinweis = screen.getByText(HINWEIS)
    expect(hinweis).toBeInTheDocument()
    expect(hinweis).toHaveClass('text-muted-foreground')
    expect(hinweis).not.toHaveClass('text-destructive')
  })

  it('wechselt bei Berührung mit ungültigem Inhalt auf destructive und zurück bei gültigem', async () => {
    const user = userEvent.setup()
    render(<RequiredHarness />)

    const textarea = screen.getByPlaceholderText('Kommentar (erforderlich)')
    await user.type(textarea, 'ab')

    const hinweis = screen.getByText(HINWEIS)
    expect(hinweis).toHaveClass('text-destructive')
    expect(hinweis).not.toHaveClass('text-muted-foreground')

    await user.type(textarea, 'c')
    expect(screen.getByText(HINWEIS)).toHaveClass('text-muted-foreground')
    expect(screen.getByText(HINWEIS)).not.toHaveClass('text-destructive')
  })

  it('zeigt bei optionalem Feld keinen Hilfetext', () => {
    render(<KommentarField onChange={vi.fn()} />)

    expect(screen.queryByText(HINWEIS)).not.toBeInTheDocument()
    expect(
      screen.getByPlaceholderText('Kommentar (optional)'),
    ).toBeInTheDocument()
  })
})
