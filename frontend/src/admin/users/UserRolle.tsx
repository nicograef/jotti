import { Badge } from '@/components/ui/badge'

import { UserRole } from './User'

// Rollen als beschriftete Badges (Design-Handoff 1e): die frühere
// Farbrätsel-Darstellung mit Stern-Symbolen wird durch ausgeschriebene Labels
// ersetzt, weil die Rolle über die Storno-Rechte entscheidet. Variante je Rolle
// laut Handoff: Admin default (primär), Serviceleitung outline, Service
// secondary.
const rolleAnzeige: Record<
  UserRole,
  { label: string; variant: 'default' | 'outline' | 'secondary' }
> = {
  [UserRole.ADMIN]: { label: 'Admin', variant: 'default' },
  [UserRole.SERVICELEITUNG]: { label: 'Serviceleitung', variant: 'outline' },
  [UserRole.SERVICE]: { label: 'Service', variant: 'secondary' },
}

export function RolleBadge({ role }: { role: UserRole }) {
  const { label, variant } = rolleAnzeige[role]
  return <Badge variant={variant}>{label}</Badge>
}
