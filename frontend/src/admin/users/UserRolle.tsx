import { Badge } from '@/components/ui/badge'

import { UserRole } from './User'

// No role-to-colour mapping: per-role badge colours are a legend-less puzzle adding nothing the label lacks.
// oxlint-disable-next-line react/only-export-components
export const rolleLabel: Record<UserRole, string> = {
  [UserRole.ADMIN]: 'Admin',
  [UserRole.SERVICELEITUNG]: 'Serviceleitung',
  [UserRole.SERVICE]: 'Service',
}

export function RolleBadge({ role }: { role: UserRole }) {
  return <Badge variant="secondary">{rolleLabel[role]}</Badge>
}
