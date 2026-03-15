import { ChevronRightIcon, Lamp } from 'lucide-react'
import { Link } from 'react-router'

import { EmptyState } from '@/components/common/EmptyState'
import { Badge } from '@/components/ui/badge'
import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'
import { formatCents } from '@/lib/utils'

import { useAktiveTische } from './table/hooks'
import { type Tisch } from './table/Tisch'

export function TableSelectionPage() {
  const { loading, tische } = useAktiveTische()

  return <>{loading ? <TischListSkeleton /> : <TischList tische={tische} />}</>
}

interface TischListComponentProps {
  tische: Tisch[]
}

function TischList(props: TischListComponentProps) {
  if (props.tische.length === 0) {
    return (
      <EmptyState
        icon={Lamp}
        title="Keine aktiven Tische"
        description="Bitte im Admin-Bereich mindestens einen Tisch aktivieren."
      />
    )
  }

  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.tische.map((tisch) => (
        <Item key={tisch.id} variant="outline" asChild>
          <Link to={`/service/tables/${tisch.id.toString()}`}>
            <ItemContent>
              <ItemTitle className="text-lg">
                <Lamp /> {tisch.name}
              </ItemTitle>
              {tisch.saldoCents < 0 && (
                <Badge variant="destructive" className="mt-1">
                  Auszahlung ausstehend:{' '}
                  {formatCents(Math.abs(tisch.saldoCents))} €
                </Badge>
              )}
            </ItemContent>
            <ItemActions>
              <ChevronRightIcon />
            </ItemActions>
          </Link>
        </Item>
      ))}
    </ItemGroup>
  )
}

function TischListSkeleton() {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {Array.from({ length: 6 }).map((_, index) => (
        <Item key={`skeleton-${index.toString()}`} variant="outline">
          <ItemContent>
            <Skeleton className="h-4 w-24" />
          </ItemContent>
          <ItemActions>
            <ChevronRightIcon />
          </ItemActions>
        </Item>
      ))}
    </ItemGroup>
  )
}
