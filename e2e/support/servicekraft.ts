import type { Locator, Page } from '@playwright/test'
import { expect } from '@playwright/test'

// Wiederverwendbare Helfer für die Servicekraft-Flows (Tischservice). Jede
// Funktion nutzt ausschließlich zugängliche Selektoren (Rolle, Platzhalter,
// Beschriftung) statt Test-IDs, passend zum Muster der Tracer-Bullet-Spec.

// zeileMit liefert die innerste Zeile (div), die sowohl den gegebenen Text als
// auch einen Button mit dem gegebenen Namen enthält. So lassen sich einzelne
// Varianten-/Positions-Zeilen ohne Test-IDs und ohne Index eindeutig treffen.
// scope ist wahlweise die ganze Seite oder ein bereits eingegrenzter Bereich
// (z. B. ein Drawer), damit sich Zeilen auch innerhalb eines Dialogs isolieren
// lassen. Der „has"-Filter wird bewusst von der Seite aus gebaut (nicht vom
// scope aus) — verkettet man ihn stattdessen vom selben scope, matcht
// Playwright das Filter-Locator nicht zuverlässig gegen die Kandidaten.
export function zeileMit(
  scope: Page | Locator,
  text: string,
  buttonName: string | RegExp,
): Locator {
  const page = 'page' in scope ? scope.page() : scope
  return scope
    .locator('div')
    .filter({ hasText: text })
    .filter({ has: page.getByRole('button', { name: buttonName }) })
    .last()
}

// oeffneHistorienDetail tippt die Historien-Zeile an, deren Beschriftung den
// gegebenen Text enthält (z. B. „Bestellung … +2,50 €"). Historien-Zeilen sind
// seit dem Redesign selbst Buttons ohne Inline-Aktionen — die ganze Zeile öffnet
// den Detail-Drawer, in dem Umbuchen/Stornieren/Drucken liegen. Der Text muss die
// Zeile eindeutig treffen (in den Specs über den Betrag).
export async function oeffneHistorienDetail(
  page: Page,
  text: string | RegExp,
): Promise<Locator> {
  await page.getByRole('button', { name: text }).click()
  const drawer = page.getByRole('dialog')
  await expect(drawer).toBeVisible()
  return drawer
}

// oeffneTisch navigiert von der Tischauswahl zur Detailseite eines Tisches
// über den „Alle Tische"-Drawer (funktioniert unabhängig davon, ob der Tisch
// bereits in „Meine Tische" markiert ist).
export async function oeffneTisch(page: Page, tisch: string): Promise<void> {
  await page.goto('/service/tische')
  await page.getByRole('button', { name: 'Alle Tische' }).click()
  await page.getByPlaceholder('Tisch suchen...').fill(tisch)
  await page
    .getByRole('button', { name: new RegExp(`^${tisch}\\b.*€`) })
    .click()
  await expect(page.getByRole('tab', { name: 'Bestellen' })).toBeVisible()
}

// Kategorie-Chips filtern die flache Variantenliste: nur die aktive Kategorie
// steht im DOM. Der erste Chip (Essen) ist per Default aktiv; Getränke- oder
// Sonstiges-Varianten erscheinen erst nach Klick auf den passenden Chip. Die
// Chips fehlen, wenn nur eine Kategorie belegt ist — daher jeder Wechsel
// tolerant (Chip nur klicken, wenn vorhanden).
const KATEGORIE_CHIPS = ['Essen', 'Getränke', 'Sonstiges']

// waehleVariante fügt auf dem Bestellen-Tab eine Variante zur gewünschten
// Menge der aktuellen Auswahl hinzu, ohne die Bestellung abzuschicken —
// Baustein für Bestellungen mit mehreren Positionen. Variantennamen („Normal",
// „Klein") kommen in mehreren Produkten vor und stehen in der flachen Liste
// alle gleichzeitig im DOM — die Zeile wird deshalb über die Produktgruppe
// (innerster Container mit dem Gruppenkopf) eingegrenzt.
export async function waehleVariante(
  page: Page,
  produkt: string,
  variante: string,
  menge = 1,
): Promise<void> {
  const gruppe = page
    .locator('div')
    .filter({ has: page.getByRole('heading', { name: produkt, exact: true }) })
    .filter({ has: page.getByRole('button', { name: 'Variante hinzufügen' }) })
    .last()
  const variantenZeile = zeileMit(gruppe, variante, 'Variante hinzufügen')
  // Flache Liste: Varianten anderer Kategorien sind nach dem Chip-Filter nicht
  // im DOM. Ist die Zeile nicht sichtbar, den passenden Kategorie-Chip
  // aktivieren — der Produktname allein verrät die Kategorie nicht, daher
  // reihum jeden vorhandenen Chip probieren, bis die Zeile erscheint.
  if (!(await variantenZeile.isVisible().catch(() => false))) {
    for (const label of KATEGORIE_CHIPS) {
      const chip = page.getByRole('button', { name: label, exact: true })
      if (await chip.isVisible().catch(() => false)) {
        await chip.click()
        if (await variantenZeile.isVisible().catch(() => false)) break
      }
    }
  }
  await expect(variantenZeile).toBeVisible()

  for (let i = 0; i < menge; i++) {
    await variantenZeile
      .getByRole('button', { name: 'Variante hinzufügen' })
      .click()
  }
}

// bestellePosition nimmt auf dem aktuell offenen Tisch eine Bestellung über
// den Bestellen-Tab auf: Variante zur gewünschten Menge hinzufügen (bei Bedarf
// über den Kategorie-Chip), Bestellung im Drawer bestätigen.
export async function bestellePosition(
  page: Page,
  produkt: string,
  variante: string,
  menge = 1,
): Promise<void> {
  await page.getByRole('tab', { name: 'Bestellen' }).click()

  await waehleVariante(page, produkt, variante, menge)

  await page.getByRole('button', { name: /Bestellung überprüfen/ }).click()
  const drawer = page.getByRole('dialog')
  await drawer.getByRole('button', { name: 'Bestellung aufnehmen' }).click()
  // .first(): Sonner kann bei mehreren Bestellungen kurzzeitig zwei Toasts mit
  // demselben Text im DOM halten (neuer Toast, alter noch beim Ausblenden).
  await expect(
    page.getByText('Bestellung wurde aufgenommen.').first(),
  ).toBeVisible()
}

// kassierePosition wechselt auf den Kassieren-Tab, wählt die Position mit dem
// angegebenen Namen zur gewünschten Menge aus und schließt die Zahlung ab.
export async function kassierePosition(
  page: Page,
  positionName: string,
  menge = 1,
): Promise<void> {
  await page.getByRole('tab', { name: 'Kassieren' }).click()
  const position = zeileMit(page, positionName, 'Produkt hinzufügen')
  await expect(position).toBeVisible()
  for (let i = 0; i < menge; i++) {
    await position.getByRole('button', { name: 'Produkt hinzufügen' }).click()
  }

  const kassierenLeiste = page.getByRole('button', { name: /Kassieren/ })
  await kassierenLeiste.click()

  const drawer = page.getByRole('dialog')
  await drawer.getByRole('button', { name: 'Kassieren' }).click()
  await expect(page.getByText('Zahlung erfolgreich.').first()).toBeVisible()
}

// offeneTischNamen liest über den „Alle Tische"-Drawer alle Tische mit einem
// Saldo ungleich 0,00 € aus. Der Kassenabschluss verlangt, dass jeder Tisch
// ausgeglichen ist — das Demo-Drehbuch des laufenden Tages lässt bewusst
// mehrere Tische in unterschiedlichen offenen Zuständen zurück (teilbezahlt,
// teilgeliefert, frisch bestellt, …). Name und Saldo stehen im DOM als eigene
// <span>-Kinder ohne Trennzeichen dazwischen (z. B. „Tisch 10,00 €" für
// „Tisch 1" mit Saldo „0,00 €") — deshalb werden sie über je einen eigenen
// <span> ausgelesen statt über den zusammengesetzten Button-Text.
async function offeneTischNamen(page: Page): Promise<string[]> {
  await page.goto('/service/tische')
  await page.getByRole('button', { name: 'Alle Tische' }).click()
  const zeilen = await page
    .getByRole('button', { name: /^.+\d,\d{2}\s*€$/ })
    .all()
  const namen: string[] = []
  for (const zeile of zeilen) {
    const spans = zeile.locator('span')
    const name = (await spans.nth(0).textContent())?.trim() ?? ''
    const saldo = (await spans.nth(1).textContent())?.trim() ?? ''
    if (name && !/^0,00\s*€$/.test(saldo)) {
      namen.push(name)
    }
  }
  await page.keyboard.press('Escape')
  return namen
}

// zeigeAlleAn klappt — falls vorhanden — die Gruppe „Von anderen" auf, die die
// Positionen anderer Servicekräfte standardmäßig eingeklappt hält. Der
// Gruppenkopf trägt die Anzahl („Von anderen · 2"); nur klicken, wenn er
// vorhanden ist (fehlt, sobald keine fremden Positionen offen sind).
async function zeigeAlleAn(page: Page): Promise<void> {
  const vonAnderenKopf = page.getByRole('button', { name: /^Von anderen ·/ })
  if (await vonAnderenKopf.isVisible().catch(() => false)) {
    await vonAnderenKopf.click()
  }
}

// vollePositionsZeilen liefert alle Positions-Zeilen (shadcn Item,
// [data-slot="item"]) mit einem „Produkt hinzufügen"-Button — das grenzt sie
// eindeutig von der umschließenden ItemGroup ab, die ebenfalls alle Buttons
// enthält.
function vollePositionsZeilen(page: Page): Locator {
  return page.locator('[data-slot="item"]').filter({
    has: page.getByRole('button', { name: 'Produkt hinzufügen' }),
  })
}

// waehleAlleVollAus klickt in jeder Positions-Zeile so oft auf „+", bis die
// Zeile voll ausgewählt ist — erkennbar an der Unterzeile „N von N ausgewählt"
// (X == Y). Anders als der „Alle auswählen"-Button, der nur eigene Positionen
// erfasst, gleicht diese Funktion jede sichtbare Zeile aus (auch fremde, sofern
// zuvor über zeigeAlleAn aufgeklappt). Eine Obergrenze pro Zeile verhindert eine
// Endlosschleife, falls die Vollauswahl-Formulierung unerwartet nie erscheint.
export async function waehleAlleVollAus(page: Page): Promise<void> {
  const zeilen = vollePositionsZeilen(page)
  const anzahlZeilen = await zeilen.count()
  for (let i = 0; i < anzahlZeilen; i++) {
    const zeile = zeilen.nth(i)
    for (let klick = 0; klick < 50; klick++) {
      const text = (await zeile.textContent()) ?? ''
      const treffer = /(\d+) von (\d+) ausgewählt/.exec(text)
      if (treffer && treffer[1] === treffer[2]) break
      await zeile.getByRole('button', { name: 'Produkt hinzufügen' }).click()
    }
  }
}

// warteAufTischGeladen wartet, bis der State-Fetch des Tisches fertig ist:
// TablePage zeigt den Header-Saldo ([data-slot="tisch-saldo"]) während des
// Ladens als „?" und danach als Euro-Betrag (z. B. „0,00 €"). Das ist ein
// deterministisches Ready-Signal — erst danach ist der Tab-Inhalt gerendert und
// Prüfungen auf Buttons/Positionszeilen lesen den fertigen DOM.
async function warteAufTischGeladen(page: Page): Promise<void> {
  await expect(page.locator('[data-slot="tisch-saldo"]')).toHaveText(
    /\d,\d{2}\s*€/,
  )
}

// settleAlleOffenenTische gleicht jeden Tisch mit offenem Saldo vollständig
// aus: alle unbezahlten Positionen kassieren. Nötig, bevor der Kassenabschluss
// angefordert werden kann (jeder Tisch muss ausgeglichen sein). Positionen
// anderer Servicekräfte stehen in der eingeklappten Gruppe „Von anderen" — die
// wird vor jeder Zählung/Auswahl aufgeklappt, sonst übersieht die Funktion
// Positionen und bricht die Schleife vorzeitig ab.
export async function settleAlleOffenenTische(page: Page): Promise<void> {
  const namen = await offeneTischNamen(page)
  for (const tisch of namen) {
    await oeffneTisch(page, tisch)

    // Auf das Fertigladen des Tisches warten, bevor auf Buttons/Zeilen geprüft
    // wird: TablePage rendert die Tab-Inhalte erst nach dem State-Fetch und zeigt
    // den Header-Saldo bis dahin als „?". Ohne dieses Ready-Signal würden die
    // isVisible()/count()-Prüfungen unten den noch leeren DOM lesen und den
    // Kassieren-Zweig stumm überspringen (Fetch-Race).
    await warteAufTischGeladen(page)

    // Alle unbezahlten Positionen kassieren (falls welche vorhanden sind). Der
    // „Kassieren"-Button ist immer im DOM (nur deaktiviert bei leerer Auswahl)
    // — ob es überhaupt unbezahlte Positionen gibt, zeigt die Positionsliste.
    await page.getByRole('tab', { name: 'Kassieren' }).click()
    await zeigeAlleAn(page)
    if ((await vollePositionsZeilen(page).count()) > 0) {
      await waehleAlleVollAus(page)
      await page.getByRole('button', { name: /Kassieren/ }).click()
      const drawer = page.getByRole('dialog')
      await drawer.getByRole('button', { name: 'Kassieren' }).click()
      await expect(
        page.getByText('Zahlung erfolgreich.').first(),
      ).toBeVisible()
    }
  }
}

// abmelden meldet den aktuell eingeloggten Benutzer über das Benutzermenü ab
// und wartet auf die Weiterleitung zur Login-Seite — zum Rollenwechsel
// innerhalb einer Spec (z. B. Service → Serviceleitung → Admin).
export async function abmelden(page: Page): Promise<void> {
  await page.getByRole('button', { name: 'Benutzermenü' }).click()
  await page.getByRole('menuitem', { name: 'Abmelden' }).click()
  await expect(page).toHaveURL(/\/login$/)
}
