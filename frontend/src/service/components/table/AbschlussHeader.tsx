import {
  DrawerDescription,
  DrawerHeader,
  DrawerTitle,
} from '@/components/ui/drawer'

interface AbschlussHeaderProps {
  // 'sheet' nutzt die Radix-Dialog-Primitive DrawerTitle/DrawerDescription (für
  // die A11y des Bottom-Sheets nötig); 'spalte' rendert eine schlichte
  // Überschrift, weil dort kein Dialog-Kontext existiert.
  variant: 'sheet' | 'spalte'
  eyebrow: string
  title: string
  // Nur im Sheet als sr-only Dialog-Description verwendet.
  description: string
}

// Gemeinsamer Kopf der drei Abschluss-Flächen (Direktverkauf, Bestellen,
// Kassieren): eine kleine Eyebrow-Zeile plus dominante Überschrift.
export function AbschlussHeader({
  variant,
  eyebrow,
  title,
  description,
}: AbschlussHeaderProps) {
  return (
    <DrawerHeader className="mx-auto w-full max-w-sm">
      <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
        {eyebrow}
      </p>
      {variant === 'sheet' ? (
        <>
          <DrawerTitle className="text-[22px] font-semibold">
            {title}
          </DrawerTitle>
          <DrawerDescription className="sr-only">
            {description}
          </DrawerDescription>
        </>
      ) : (
        <h2 className="font-heading text-[22px] font-semibold text-foreground">
          {title}
        </h2>
      )}
    </DrawerHeader>
  )
}
