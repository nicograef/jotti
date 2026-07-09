// Gemeinsame Layout-Bausteine der Admin-Listen (Produkte, Tische, Benutzer).

// Unterer Freiraum, damit die letzte Karte beim Scrollen vollständig über dem
// fixierten FAB (New*Dialog) sichtbar und bedienbar bleibt. Der FAB steht bei
// bottom = 1rem + Safe-Area und ist 36px hoch; der Freiraum deckt FAB-Höhe,
// Safe-Area und einen sichtbaren Abstand ab. An genau einer Stelle definiert
// und von allen drei Admin-Listen genutzt.
export const adminListBottomClearance =
  'pb-[calc(6rem+env(safe-area-inset-bottom,0px))]'

// Touch-Ziel für die Zeilen-Aktionen (Bearbeiten/Löschen): mindestens 44x44px.
export const adminItemActionButton = 'size-11'
