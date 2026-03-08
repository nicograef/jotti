import { ChevronRightIcon, Lamp } from 'lucide-react'
import { Link } from 'react-router'

import {
  Item,
  ItemActions,
  ItemContent,
  ItemGroup,
  ItemTitle,
} from '@/components/ui/item'
import { Skeleton } from '@/components/ui/skeleton'

import { useAktiveTische } from './table/hooks'
import { type Tisch } from './table/Tisch'

export function TableSelectionPage() {
  const { loading, tische } = useAktiveTische()

  return (
    <>
      <h1 className="text-2xl font-bold">Tisch auswählen</h1>
      {loading ? <TischListSkeleton /> : <TischList tische={tische} />}
    </>
  )
}

interface TischListComponentProps {
  tische: Tisch[]
}

function TischList(props: TischListComponentProps) {
  return (
    <ItemGroup className="grid gap-2 lg:grid-cols-2 2xl:grid-cols-3 my-4">
      {props.tische.map((tisch) => (
        <Item key={tisch.id} variant="outline" asChild>
          <Link to={`/service/tables/${tisch.id.toString()}`}>
            <ItemContent>
              <ItemTitle className="text-lg">
                <Lamp /> {tisch.name}
              </ItemTitle>
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
