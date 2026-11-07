import { describe, expect, it } from 'vitest'

import type { Produkt } from '@/lib/produktSchemas'
import { testProdukt } from '@/test/fixtures'

import {
  mitPfand,
  type PfandSchritt,
  pfandVerknuepfungen,
  type PfandZiel,
} from './pfandKopplung'

// Bier 0,5 l (1) und Bier 0,3 l (2) ziehen das Glaspfand (19) mit, Wasser (3) nichts.
const verknuepfungen = new Map<number, PfandZiel>([
  [1, { varianteId: 19, preisCents: 200 }],
  [2, { varianteId: 19, preisCents: 200 }],
])

describe('mitPfand', () => {
  it.each<{
    fall: string
    vorher: Record<number, number>
    varianteId: number
    schritt: PfandSchritt
    nachher: Record<number, number>
  }>([
    {
      fall: 'Plus legt das Pfand dazu',
      vorher: {},
      varianteId: 1,
      schritt: 'plus',
      nachher: { 1: 1, 19: 1 },
    },
    {
      fall: 'Minus nimmt das Pfand mit',
      vorher: { 1: 2, 19: 2 },
      varianteId: 1,
      schritt: 'minus',
      nachher: { 1: 1, 19: 1 },
    },
    {
      fall: 'Minus an einer Variante auf null ändert nichts',
      vorher: { 19: 1 },
      varianteId: 1,
      schritt: 'minus',
      nachher: { 19: 1 },
    },
    {
      fall: 'Minus senkt das Pfand nicht unter null',
      vorher: { 1: 2, 19: 0 },
      varianteId: 1,
      schritt: 'minus',
      nachher: { 1: 1, 19: 0 },
    },
    {
      fall: 'zwei Varianten mit einem Pfand summieren sich',
      vorher: { 1: 1, 19: 1 },
      varianteId: 2,
      schritt: 'plus',
      nachher: { 1: 1, 2: 1, 19: 2 },
    },
    {
      fall: 'Plus am Pfand ändert nur das Pfand',
      vorher: { 1: 1, 19: 1 },
      varianteId: 19,
      schritt: 'plus',
      nachher: { 1: 1, 19: 2 },
    },
    {
      fall: 'Minus am Pfand ändert nur das Pfand',
      vorher: { 1: 1, 19: 1 },
      varianteId: 19,
      schritt: 'minus',
      nachher: { 1: 1, 19: 0 },
    },
    {
      fall: 'eine Variante ohne Verknüpfung zählt allein',
      vorher: {},
      varianteId: 3,
      schritt: 'plus',
      nachher: { 3: 1 },
    },
  ])('$fall', ({ vorher, varianteId, schritt, nachher }) => {
    expect(mitPfand(vorher, varianteId, schritt, verknuepfungen)).toEqual(
      nachher,
    )
  })
})

describe('pfandVerknuepfungen', () => {
  const bier: Produkt = {
    ...testProdukt,
    id: 2,
    name: 'Bier',
    varianten: [
      { ...testProdukt.varianten[0], id: 1, pfandVarianteId: 19 },
      { ...testProdukt.varianten[0], id: 2, pfandVarianteId: 29 },
      { ...testProdukt.varianten[0], id: 3, pfandVarianteId: null },
    ],
  }
  const pfand: Produkt = {
    ...testProdukt,
    id: 9,
    name: 'Pfand',
    pfand: true,
    varianten: [{ ...testProdukt.varianten[0], id: 19, preisCents: 200 }],
  }

  it('nimmt nur Verweise auf, deren Ziel unter den Produkten steht', () => {
    expect(pfandVerknuepfungen([bier, pfand])).toEqual(
      new Map([[1, { varianteId: 19, preisCents: 200 }]]),
    )
  })
})
