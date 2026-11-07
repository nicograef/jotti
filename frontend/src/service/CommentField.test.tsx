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

  // Das Backend zählt den Kommentar in Codepoints (max. 100); der Cursor zählt UTF-16-Einheiten.
  // Gekürzt wird nur die Einfügung, der bestehende Text und der Cursor bleiben.
  it.each([
    ['', 0, 'ä'.repeat(101), 'ä'.repeat(100), 100],
    ['', 0, 'a'.repeat(99) + '😀😀', 'a'.repeat(99) + '😀', 101],
    ['a'.repeat(100), 50, 'x', 'a'.repeat(100), 50],
    ['a'.repeat(98), 49, 'xyz', 'a'.repeat(49) + 'xy' + 'a'.repeat(49), 51],
    ['ä'.repeat(99), 0, 'öö', 'ö' + 'ä'.repeat(99), 1],
  ])(
    'kürzt die Einfügung auf 100 Zeichen an einer Zeichengrenze (Fall %#)',
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
