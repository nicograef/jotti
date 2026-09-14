// Dekorativer Glow hinter dem AdminPageHeader. Der äußere Container clippt den
// Überhang (overflow-hidden), sonst entsteht ein horizontaler Scrollbalken.
export type SpektralFarbe =
  'red' | 'orange' | 'green' | 'teal' | 'blue' | 'violet'

export function HeaderGlow({
  farben = ['teal', 'violet'],
}: {
  farben?: readonly [SpektralFarbe, SpektralFarbe]
}) {
  const [erste, zweite] = farben
  return (
    <div
      aria-hidden
      className="pointer-events-none absolute inset-0 -z-10 overflow-hidden print:hidden"
    >
      <div
        className="absolute -top-[70px] -left-10 h-[200px] w-[460px] opacity-[0.18] blur-[52px]"
        style={{
          background: `radial-gradient(ellipse at 20% 40%, color-mix(in oklab, var(--sp-${erste}) 40%, transparent), transparent 60%), radial-gradient(ellipse at 70% 60%, color-mix(in oklab, var(--sp-${zweite}) 28%, transparent), transparent 60%)`,
        }}
      />
    </div>
  )
}
