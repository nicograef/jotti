import { z } from 'zod'

export { DateStringSchema } from '@/lib/utils'

export const SteuersatzSchema = z.enum([
  'regel',
  'ermaessigt',
  'befreit',
  'kombi',
])
