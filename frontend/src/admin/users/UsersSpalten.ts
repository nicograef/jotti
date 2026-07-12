// Gemeinsames Spaltenraster für die Benutzertabelle (Design-Handoff 1e):
// Name·Login, Rolle, Status, Aktionen. Header (Users.tsx) und Zeilen
// (UserRow.tsx) müssen dieselbe Konstante verwenden, sonst driften die
// Spaltenbreiten auseinander (beide sind unabhängige Grid-Container,
// `fr`/`auto` lösen sonst pro Container auf).
export const BENUTZER_SPALTEN =
  'grid grid-cols-[1.4fr_1fr_0.8fr_96px] items-center gap-x-3'
