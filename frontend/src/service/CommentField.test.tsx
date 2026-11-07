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

  // Das Backend zählt den Kommentar in UTF-8-Bytes (max. 100), nicht in Zeichen.
  // Gekürzt wird nur die Einfügung, der bestehende Text und der Cursor bleiben.
  it.each([
    ['', 0, 'ä'.repeat(60), 'ä'.repeat(50), 50],
    ['', 0, 'a'.repeat(99) + 'ä', 'a'.repeat(99), 99],
    ['', 0, 'a'.repeat(97) + '😀', 'a'.repeat(97), 97],
    ['', 0, 'a'.repeat(96) + '😀', 'a'.repeat(96) + '😀', 98],
    ['a'.repeat(100), 50, 'x', 'a'.repeat(100), 50],
    ['a'.repeat(98), 49, 'xyz', 'a'.repeat(49) + 'xy' + 'a'.repeat(49), 51],
  ])(
    'kürzt die Einfügung auf 100 Bytes an einer Zeichengrenze (Fall %#)',
    async (vorhanden, cursor, eingabe, erwartet, cursorDanach) => {
      const user = userEvent.setup()
      const onChange = vi.fn()
      render(<KommentarField onChange={onChange} />)

      const textarea = screen.getByPlaceholderText<HTMLTextAreaElement>(
        'Kommentar (optional)',
      )
      await user.click(textarea)
      if (vorhanden) {
        await user.paste(vorhanden)
      }
      textarea.setSelectionRange(cursor, cursor)
      await user.paste(eingabe)

      expect(textarea).toHaveValue(erwartet)
      expect(onChange).toHaveBeenLastCalledWith(erwartet)
      expect(textarea.selectionStart).toBe(cursorDanach)
    },
  )

  it('zeigt bei optionalem Feld keinen Hilfetext', () => {
    render(<KommentarField onChange={vi.fn()} />)

    expect(screen.queryByText(HINWEIS)).not.toBeInTheDocument()
    expect(
      screen.getByPlaceholderText('Kommentar (optional)'),
    ).toBeInTheDocument()
  })
})
