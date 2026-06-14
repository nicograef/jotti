import { z } from 'zod'

export const DateStringSchema = z
  .string()
  .refine((date) => !isNaN(Date.parse(date)), {
    message: 'Ungültiges Datumsformat',
  })

export const SteuersatzSchema = z.enum([
  'regel',
  'ermaessigt',
  'befreit',
  'kombi',
])
