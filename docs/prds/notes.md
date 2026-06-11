# Notizen aus erstem Usability-Testing

Datum: 10. Juni 2026, Vereinsheim, Ubuntu + Docker Compose Local-Stack, 2 Bondrucker (WLAN + LAN), Smartphones im Vereins-WLAN.

## Drucker-Relay Queue-Blocking

**Problem:** Das Relay verarbeitet alle Druckaufträge sequenziell in einer einzigen Schleife. Wenn ein Drucker nicht erreichbar ist (z.B. falsches Subnetz, Drucker aus), blockiert das Relay mit bis zu 60 Retries × 5s = 5 Minuten pro Auftrag — und in dieser Zeit werden auch Aufträge für andere, erreichbare Drucker nicht verarbeitet.

**Beobachtet:** LAN-Drucker (192.168.1.100) war in anderem Subnetz als WLAN (192.168.178.x). Relay hing in Endlos-Retries fest, WLAN-Drucker (192.168.178.135) wurde nicht bedient. Workaround: Manuelle DB-Manipulation (`UPDATE druckauftraege SET status='gedruckt'`).

**Gewünschte Lösung:**

- Drucker unabhängig voneinander bedienen (concurrent per ZielIP).
- Retry-Limit deutlich reduzieren (z.B. 3 statt 60).
- Bei Fehlschlag: Auftrag als `fehlgeschlagen` markieren (neuer Status), nicht endlos wiederholen.
- Admin-Oberfläche/Endpoint zum Verwalten fehlgeschlagener Druckaufträge (erneut versuchen, verwerfen).
- Kein manueller DB-Eingriff nötig im Betrieb.
