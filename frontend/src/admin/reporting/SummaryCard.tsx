import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { useCountUp } from '@/hooks/use-count-up'
import { formatCents } from '@/lib/utils'

export function SummaryCard({
  title,
  valueCents,
  sub,
}: {
  title: string
  valueCents: number
  sub?: string
}) {
  const wert = useCountUp(valueCents)
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-sm text-muted-foreground">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-xl font-bold tabular-nums">{formatCents(wert)} €</p>
        {sub && <p className="mt-0.5 text-sm text-muted-foreground">{sub}</p>}
      </CardContent>
    </Card>
  )
}
