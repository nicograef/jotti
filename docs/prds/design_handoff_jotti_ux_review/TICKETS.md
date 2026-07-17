# Tickets für Claude Code

Jeder Block ist ein eigenständiger Prompt. Repo: `nicograef/jotti`, alle Pfade relativ zu `frontend/src/`.

---

## Block A — vor 1.0

### A1 · Warenkorb überlebt Tab-Wechsel nicht (F1, Impact hoch)

```
Im Service-Bereich geht die Auswahl beim Tab-Wechsel verloren: Radix-Tabs unmounten
inaktive TabsContent (components/ui/tabs.tsx, bewusst wegen fadeUp), aber die
useMengen-States leben lokal in den Tab-Flächen.

Betroffen:
- service/components/table/Bestellung.tsx (useMengen für den Bestell-Korb)
- service/components/table/Zahlung.tsx (useMengen für die Kassieren-Auswahl)
- Host: service/TablePage.tsx

Aufgabe: Hebe die beiden useMengen-Instanzen (inkl. zugehörigem UI-State wie
andereOffen NICHT — nur die Mengen) nach TablePage und reiche sie als Props an
Bestellung und Zahlung durch. Der bestellungId-/Eingabe-Lebenszyklus in
BestellungAbschluss/ZahlungAbschluss (warLeerRef-Muster) darf sich nicht ändern.
Kein forceMount; die Tab-fadeUp-Animation bleibt.

Akzeptanz:
- 390 px: 3 Positionen wählen → Tab Historie → zurück zu Bestellen → Auswahl und
  Dock-Button-Summe unverändert. Gleiches für Kassieren.
- Nach erfolgreichem Bestellen/Kassieren wird die jeweilige Auswahl weiterhin geleert.
- ≥1024 px (Split-Layout) verhält sich unverändert.
- TablePage.test.tsx / Bestellung.test.tsx / Zahlung.test.tsx grün; neuen Test für
  „Auswahl überlebt Tab-Wechsel" ergänzen.
```

### A2 · Storno/Umbuchung ohne konsistentes Erfolgs-Feedback (F4)

```
Erfolgs-Feedback vereinheitlichen: Bestellen/Kassieren/Direktverkauf nutzen den
ErfolgsPop (service/components/ErfolgsPop.tsx), Umbuchung nutzt toast.success,
Stornierung hat GAR KEIN Erfolgs-Feedback.

- service/components/table/HistorieStornierungDrawer.tsx: onStornierungErteilt
  schließt nur den Drawer.
- service/components/table/HistorieUmbuchungDrawer.tsx: toast.success('Bestellung umgebucht.')

Aufgabe: Beide Flows an den bestehenden ErfolgsPop-Mechanismus in TablePage
anschließen (zeigeErfolg wird bereits an Bestellung/Zahlung gereicht — analog über
TischHistorie an die beiden Drawer durchreichen). Texte: „Stornierung gebucht." und
„Auf {Zielname} umgebucht." Der bisherige Sofort-Refetch der Drawer folgt dem
Pop-Muster: Refetch erst beim Schließen des Pops (wie bei Bestellen/Kassieren).
toast.success in HistorieUmbuchungDrawer entfernen.

Akzeptanz: Nach Storno und Umbuchung erscheint der Vollbild-Pop; nach dem
Ausblenden sind Historie/Saldo aktualisiert. Bestehende Drawer-Tests angepasst.
```

### A3 · EuroInput schluckt Ziffern (Debounce-Reformat)

```
components/common/EuroInput.tsx normalisiert per 1-s-Debounce mitten im Tippen:
onValueChange(formatBlur(cleaned)) nach Pause. Wer langsam „15" tippt, hat nach 1 s
„1,00" im Feld; die nächste Ziffer wird von cleanInput als 3. Nachkommastelle
verworfen.

Aufgabe: Debounce-Reformat komplett entfernen; Normalisierung nur noch onBlur.
debounceRef, den Cleanup-Effect und den setTimeout-Zweig löschen.

Akzeptanz: Langsame Eingabe „1" [Pause >1s] „5" ergibt „15". Blur formatiert zu
„15,00". EuroInput.test.tsx erweitert um den Pause-Fall.
```

### A4 · Toter Anmelden-Button

```
components/common/LoginForm.tsx (und PasswordForm.tsx): Submit-Button ist
disabled={loading || !form.formState.isValid} bei mode:'onSubmit' — ein Tap auf den
gesperrten Button löst keine Validierung aus, der Nutzer sieht nie warum.

Aufgabe: disabled nur noch an `loading` koppeln; Klick löst handleSubmit aus und
zeigt die Feldfehler (RHF macht das bei onSubmit bereits). PasswordForm analog
(dort mode:'onTouched' — isValid-Kopplung ebenfalls entfernen, loading reicht).

Akzeptanz: Leeres Formular + Klick auf „Anmelden" zeigt Feldfehler; Doppel-Submit
weiterhin durch loading verhindert. Bestehende Tests grün.
```

---

## Block B — Copy-&-Token-Sweep (½ Tag)

### B1 · Button-Variant `destructive-solid` (F5, A11y Dark Mode)

```
Sechs AlertDialog-Bestätigungen stylen ad-hoc: className="bg-destructive text-white"
(products/ProductItem.tsx, products/EditVariantDialog.tsx, tables/EditTischDialog.tsx,
users/UserRow.tsx, settings/DruckstationConfigPage.tsx — per Grep verifizieren).
Im Dark Mode ist --destructive red-400 → Weiß darauf ≈ 2,9:1, unter WCAG AA.

Aufgabe: In components/ui/button.tsx einen Variant `destructive-solid` nach dem
warn-Muster (ADR 04) anlegen: solide destruktive Fläche, die ihren Kontrast selbst
trägt — Light: bg-destructive + weißer Text (Light-destructive ist dunkel genug),
Dark: helle Fläche + dunkler Text (neues Token --destructive-solid-foreground in
index.css, Dark ≈ red-950). Alle Call-Sites auf den Variant umstellen, Ad-hoc-
Klassen entfernen. e2e/tests/admin-kontrast-axe.spec.ts muss die Dialoge abdecken.

Akzeptanz: Kein `bg-destructive text-white` mehr im Repo; axe-Kontrast-Spec grün
in Light und Dark.
```

### B2 · Copy- und Icon-Fixes (ein Commit)

```
Kleine Konsistenz-Fixes, keine Verhaltensänderung. E2E-/Unit-Tests matchen auf
Texte — betroffene Assertions mit anpassen.

1. lib/utils.ts formatAlleAuswaehlenLabel: Im Kassieren-Kontext wählt der Button nur
   EIGENE Positionen → Label „Meine N Positionen auswählen · X €" (Parameter oder
   zweite Funktion; die Umbuchung behält „Alle", dort stimmt es).
2. admin/reporting/LiveReportingSection.tsx: Refresh-Button „Jetzt" → „Aktualisieren".
3. admin/products/ProductItem.tsx: Dropdown „Umbenennen" → „Bearbeiten" (gleiches
   Label wie der Stift-Tooltip; beide öffnen denselben Dialog).
4. admin/AdminSidebar.tsx: Icon für „Zum Service-Bereich" von LogOut auf
   ArrowRightLeft ändern (LogOut bleibt exklusiv beim Abmelden).
5. admin/kasse/KasseAbschliessenSection.tsx: Differenz-Anzeige immer mit Vorzeichen
   („+12,50 €" bei Überschuss); Formatierung als kleine Helper-Fn neben formatEuro.
6. service/TablePage.tsx: Ladezustand „Tisch ??" / „?" durch Skeleton-Bausteine
   ersetzen (components/ui/skeleton.tsx, wie TischListSkeleton).
```

### B3 · „Unbezahlt" ist kein Gefahrenzustand

```
Rot ist per ADR 04 für gefährlich/destruktiv reserviert; der Normalzustand eines
offenen Tischs trägt es trotzdem:
- service/TablePage.tsx: Badge variant="destructive" „N unbezahlt"
- service/components/MeinTischCard.tsx: statusFarbe bg-destructive bei eigenen
  offenen Positionen

Aufgabe: Badge auf neutrale/amber Darstellung umstellen (z. B. Badge secondary oder
amber-Tint analog --warn-Fläche, aber als Soft-Tint); MeinTischCard-Punkt: eigene
offene → amber, fremde offene → neutral, erledigt → grün (heute: rot/amber/grün).
Rot bleibt Storno-Beträgen und Fehlerzuständen vorbehalten.

Akzeptanz: Kein bg-destructive/variant-destructive mehr für „unbezahlt"-Zustände;
Stornierungs-Rot unverändert. Screens in Light+Dark geprüft.
```

---

## Block C — nach 1.0

### C1 · Mode-Switcher im Service-Header (F3)

```
Der Wechsel Tischservice ↔ Direktverkauf liegt nur im UserDropdown
(components/common/UserDropdown.tsx, moduswechselEintrag). Routen, Loader und
Persistenz existieren (lib/arbeitsmodus.ts, routes.ts).

Aufgabe: In service/ServiceLayout.tsx den Titel durch eine Segmented Control
„Tische | Theke" ersetzen (TabsList-Optik, aber Navigation via NavLink auf
/service/tische bzw. /service/direktverkauf; aktiver Zustand aus useMatch). Auf der
Tisch-Detailseite bleibt der „‹ Meine Tische"-Backlink statt des Switchers.
Menüeintrag im UserDropdown bleibt als Zweitweg. Den Hinweistext im EmptyState von
TableSelectionPage („…im Benutzermenü oben rechts") entsprechend kürzen.

Akzeptanz: Wechsel in beide Richtungen per 1 Tap ab jeder Modus-Startseite;
Arbeitsmodus-Persistenz (zuletzt genutzter Modus nach Login) unverändert;
44-px-Zielgröße; Tests für ServiceLayout/TableSelectionPage angepasst.
```

### C2 · Aufrunden-Chips statt „Zahlbetrag inkl. Trinkgeld" (F2)

```
service/components/table/ZahlungAbschluss.tsx: Das Zielbetrag-Feld + 3-Zeilen-Hinweis
durch Aufrunden-Chips ersetzen; Direktverkauf (DirektverkaufAbschluss.tsx) bekommt
dieselbe Komponente (hat heute gar keine Trinkgeld-Option).

Aufgabe: Neue Komponente service/components/table/AufrundenChips.tsx:
- Chips aus totalCents abgeleitet: exakter Betrag („13 € genau"), dann die nächsten
  2–3 glatten Beträge (aufgerundet auf 1 €, 5 €; Duplikate entfernen), plus
  „Anderer…", der das bisherige EuroInput-Feld einblendet.
- Auswahl setzt zielbetragEuro; Trinkgeld- und Rückgeld-Zeilen wie bisher aus
  calculateZahlungsbetraege. Chip abwählbar (Tap auf aktiven Chip = zurück zu genau).
- Keine Backend-/Nutzlast-Änderung: Zielbetrag/Erhalten gehen weiterhin NICHT an
  die API (nur Anzeige-Rechnung, siehe ADR 08).
- Chips ≥44 px hoch, tabular-nums.

Akzeptanz: „machen wir 15" = 1 Tap; Hinweistext entfällt; freies Feld weiter
erreichbar; Sheet (<1024) und Spalte (≥1024) identisch; Zahlung/Direktverkauf-Tests
erweitert.
```

### C3 · Kleinere Flow-Verbesserungen

```
1. TischAuswahlDrawer.tsx („Alle Tische"): Suchfeld oben im Drawer (gleiche
   Filterlogik wie die Hauptsuche in TableSelectionPage) ODER Sortierung im Drawer
   nach Name statt Saldo — eines von beiden, damit gezieltes Finden ohne Schließen
   möglich ist.
2. TischHistorie.tsx: Für Rollen mit canCancel eine direkte „Stornieren…"-Aktion an
   der Historien-Zeile (z. B. Icon-Button rechts), Detail-Drawer bleibt Standardweg.
3. DruckstationConfigPage.tsx: Nach on-blur-Speichern der Drucker-IP kurzes
   Inline-„Gespeichert ✓" am Feld (2 s), zusätzlich zum Toast.
```
