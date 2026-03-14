import {
  ClipboardList,
  LayoutDashboard,
  Receipt,
  RefreshCw,
  ShoppingCart,
  TriangleAlert,
} from 'lucide-react'
import { Link } from 'react-router'

import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Item,
  ItemContent,
  ItemDescription,
  ItemGroup,
  ItemMedia,
  ItemTitle,
} from '@/components/ui/item'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCents } from '@/lib/utils'

import { useDashboard } from '../reporting/hooks'

function KpiCard({
  title,
  value,
  sub,
  icon,
}: {
  title: string
  value: string
  sub?: string
  icon: React.ReactNode
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-sm text-muted-foreground">
          {icon}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-2xl font-bold">{value}</p>
        {sub && <p className="mt-0.5 text-sm text-muted-foreground">{sub}</p>}
      </CardContent>
    </Card>
  )
}

function KpiSkeleton() {
  return (
    <Card>
      <CardHeader>
        <Skeleton className="h-4 w-24" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-8 w-20" />
        <Skeleton className="mt-1.5 h-3.5 w-16" />
      </CardContent>
    </Card>
  )
}

export function AdminDashboardPage() {
  const { data, loading } = useDashboard()

  return (
    <>
      <div className="flex items-center gap-3">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <Badge variant="secondary" className="gap-1.5">
          <RefreshCw className="size-3" />
          Live
        </Badge>
      </div>

      <div className="mt-4 grid grid-cols-2 gap-4 lg:grid-cols-4">
        {loading ? (
          <>
            <KpiSkeleton />
            <KpiSkeleton />
            <KpiSkeleton />
            <KpiSkeleton />
          </>
        ) : (
          <>
            <KpiCard
              title="Gesamtumsatz"
              value={`${formatCents(data.gesamtUmsatzCents)} €`}
              icon={<Receipt className="h-4 w-4" />}
            />
            <KpiCard
              title="Offene Tische"
              value={String(data.anzahlOffeneTische)}
              icon={<LayoutDashboard className="h-4 w-4" />}
            />
            <KpiCard
              title="Bestellungen"
              value={String(data.anzahlBestellungen)}
              sub={`${formatCents(data.gesamtBestellungenCents)} €`}
              icon={<ShoppingCart className="h-4 w-4" />}
            />
            <KpiCard
              title="Stornierungen"
              value={String(data.anzahlStornierungen)}
              sub={`${formatCents(data.gesamtStornierungenCents)} €`}
              icon={<TriangleAlert className="h-4 w-4" />}
            />
          </>
        )}
      </div>

      <Separator className="mt-8" />

      <h2 className="mt-6 text-lg font-semibold">Schnellzugriff</h2>
      <ItemGroup className="mt-3">
        <Item asChild variant="outline" size="sm">
          <Link to="/admin/tagesabrechnung">
            <ItemMedia variant="icon">
              <ClipboardList />
            </ItemMedia>
            <ItemContent>
              <ItemTitle>Tagesabrechnung</ItemTitle>
              <ItemDescription>
                Umsatzübersicht für einen Zeitraum
              </ItemDescription>
            </ItemContent>
          </Link>
        </Item>
      </ItemGroup>
    </>
  )
}
