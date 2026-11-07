import { formatEuro } from '@/lib/utils'

// betrag ist in Cent.
export function GesamtZeile({
  label,
  betrag,
}: {
  label: string
  betrag: number
}) {
  return (
    <div className="flex justify-between border-t-2 pt-2 font-bold">
      <div>{label}</div>
      <div>{formatEuro(betrag)}</div>
    </div>
  )
}
