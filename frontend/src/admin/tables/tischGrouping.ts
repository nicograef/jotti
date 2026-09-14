import { type Tisch, WEITERE_GRUPPE } from './Tisch'

export interface TischGruppe {
  name: string
  tische: Tisch[]
}

// 'Zelt 3' → { praefix: 'Zelt', nummer: 3 }; ohne abschließende Zahl gibt es
// keinen Präfix (nummer null).
function zerlegeNamen(name: string): {
  praefix: string
  nummer: number | null
} {
  const match = /^(.*?)(\d+)\s*$/.exec(name)
  if (!match) {
    return { praefix: '', nummer: null }
  }
  return { praefix: match[1].trim(), nummer: Number(match[2]) }
}

// Tische ohne abschließende Zahl — oder mit leerem Präfix wie '12' — landen in
// der Gruppe 'Weitere', die immer zuletzt steht. Innerhalb einer Gruppe
// numerisch nach der abschließenden Zahl sortiert, die Gruppen nach erstem
// Auftreten in der Eingabe (Backend liefert nach ID aufsteigend).
export function gruppiereTische(tische: Tisch[]): TischGruppe[] {
  const gruppen = new Map<string, Tisch[]>()

  for (const tisch of tische) {
    const { praefix, nummer } = zerlegeNamen(tisch.name)
    const gruppenName =
      praefix !== '' && nummer !== null ? praefix : WEITERE_GRUPPE
    const bestehend = gruppen.get(gruppenName)
    if (bestehend) {
      bestehend.push(tisch)
    } else {
      gruppen.set(gruppenName, [tisch])
    }
  }

  const result: TischGruppe[] = []
  for (const [name, gruppenTische] of gruppen) {
    if (name === WEITERE_GRUPPE) {
      continue
    }
    result.push({
      name,
      tische: [...gruppenTische].sort(
        (a, b) =>
          (zerlegeNamen(a.name).nummer ?? 0) -
          (zerlegeNamen(b.name).nummer ?? 0),
      ),
    })
  }

  const weitere = gruppen.get(WEITERE_GRUPPE)
  if (weitere) {
    result.push({ name: WEITERE_GRUPPE, tische: weitere })
  }

  return result
}
