# Notizen aus erstem Usability-Testing

Datum: 10. Juni 2026, Vereinsheim, Ubuntu + Docker Compose Local-Stack, 2 Bondrucker (WLAN + LAN), Smartphones im Vereins-WLAN.

## Fehlermeldungen / UX

- Kassensitzung konnte aufgrund fehlender Betreiberinformationen nicht geöffnet werden, aber es gab keine hilfreiche Fehlermeldung für den User. Der Fehler kam als generischer Server-Error zurück, statt auf die fehlende Konfiguration hinzuweisen.

## Drucker-Relay Queue-Blocking

**Problem:** Das Relay verarbeitet alle Druckaufträge sequenziell in einer einzigen Schleife. Wenn ein Drucker nicht erreichbar ist (z.B. falsches Subnetz, Drucker aus), blockiert das Relay mit bis zu 60 Retries × 5s = 5 Minuten pro Auftrag — und in dieser Zeit werden auch Aufträge für andere, erreichbare Drucker nicht verarbeitet.

**Beobachtet:** LAN-Drucker (192.168.1.100) war in anderem Subnetz als WLAN (192.168.178.x). Relay hing in Endlos-Retries fest, WLAN-Drucker (192.168.178.135) wurde nicht bedient. Workaround: Manuelle DB-Manipulation (`UPDATE druckauftraege SET status='gedruckt'`).

**Gewünschte Lösung:**

- Drucker unabhängig voneinander bedienen (concurrent per ZielIP).
- Retry-Limit deutlich reduzieren (z.B. 3 statt 60).
- Bei Fehlschlag: Auftrag als `fehlgeschlagen` markieren (neuer Status), nicht endlos wiederholen.
- Admin-Oberfläche/Endpoint zum Verwalten fehlgeschlagener Druckaufträge (erneut versuchen, verwerfen).
- Kein manueller DB-Eingriff nötig im Betrieb.

## Setup / Infrastruktur

- `init-env.sh` generiert Base64-Passwörter, die URL-unsichere Zeichen enthalten (`/`, `+`, `=`). Das bricht den Connection-String im Migrate-Container. Workaround: Hex-Passwort manuell gesetzt. Script sollte `openssl rand -hex 32` verwenden.
- Alpine-Image hat kein `tzdata` — Zeitzone-Lookup schlug fehl. Gelöst mit `import _ "time/tzdata"` im Backend (bereits committed).
- Relay braucht `RELAY_TLS_SKIP_VERIFY=1` für selbstsignierte Zertifikate im Local-Stack.
