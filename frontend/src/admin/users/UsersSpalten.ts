// Header (Users.tsx) und Zeilen (UserRow.tsx) müssen dieselbe Konstante
// verwenden: es sind unabhängige Grid-Container, `fr`/`auto` lösen sonst pro
// Container auf und die Spaltenbreiten driften auseinander.
export const BENUTZER_SPALTEN =
  'grid grid-cols-[1.4fr_1fr_0.8fr_96px] items-center gap-x-3'
