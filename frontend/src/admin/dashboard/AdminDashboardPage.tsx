import {
  ClipboardList,
  LayoutDashboard,
  Receipt,
  ShoppingCart,
  TriangleAlert,
} from 'lucide-react'
import { Link } from 'react-router'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCents } from '@/lib/utils'

import { useDashboard } from '../reporting/hooks'

function KpiCard({
  title,
  value,
  icon,
}: {
  title: string
  value: string
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
      </CardContent>
    </Card>
  )
}

export function AdminDashboardPage() {
  const { data, loading } = useDashboard()

  return (
    <>
      <h1 className="text-2xl font-bold">Dashboard</h1>

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
              icon={<ShoppingCart className="h-4 w-4" />}
            />
            <KpiCard
              title="Stornierungen"
              value={String(data.anzahlStornierungen)}
              icon={<TriangleAlert className="h-4 w-4" />}
            />
          </>
        )}
      </div>

      <h2 className="mt-8 text-lg font-semibold">Schnellzugriff</h2>
      <div className="mt-2 flex gap-4">
        <Link
          to="/admin/tagesabrechnung"
          className="flex items-center gap-2 rounded-md border px-4 py-2 text-sm hover:bg-accent"
        >
          <ClipboardList className="h-4 w-4" />
          Tagesabrechnung
        </Link>
      </div>
    </>
  )
}
