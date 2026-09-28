/**
 * An object instead of calling `window.location.reload()` in place: jsdom does
 * not let tests replace that method, but `vi.spyOn(Seite, 'neuLaden')` works.
 */
export const Seite = {
  neuLaden: (): void => {
    window.location.reload()
  },
}
