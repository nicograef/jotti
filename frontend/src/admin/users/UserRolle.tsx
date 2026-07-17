import { Badge } from '@/components/ui/badge'

import { UserRole } from './User'

// Rollen als beschriftete Badges: die Rolle wird ausschließlich über das
// ausgeschriebene Label unterschieden, alle Rollen teilen denselben neutralen
// Badge-Variant. Das kehrt die frühere Drei-Varianten-Absicht („Design-Handoff
// 1e": Admin default, Serviceleitung outline, Service secondary) bewusst um —
// unterschiedliche Badge-Farben pro Rolle sind ein Farbrätsel ohne Legende und
// tragen keine Bedeutung, die der Text nicht schon eindeutig trägt.
// eslint-disable-next-line react-refresh/only-export-components
export const rolleLabel: Record<UserRole, string> = {
  [UserRole.ADMIN]: 'Admin',
  [UserRole.SERVICELEITUNG]: 'Serviceleitung',
  [UserRole.SERVICE]: 'Service',
}

export function RolleBadge({ role }: { role: UserRole }) {
  return <Badge variant="secondary">{rolleLabel[role]}</Badge>
}
