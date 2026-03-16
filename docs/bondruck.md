# Bondruck

jotti kann Bestellungen automatisch als Bons an Küche und Getränketheke drucken. Sobald eine Servicekraft eine Bestellung aufnimmt, werden die Bons innerhalb weniger Sekunden an den richtigen Drucker gesendet — Essenspositionen an den Küchendrucker, Getränke an den Thekendrucker.

---

## Funktionsweise

```
Servicekraft nimmt Bestellung auf (Smartphone)
        │
        ▼
jotti speichert die Bestellung (Cloud-Server)
        │
        ▼
Print-Relay holt neue Bestellungen ab (alle 2 Sekunden)
        │
        ▼
Bons werden automatisch gedruckt (Küche / Theke)
```

**Zwei Komponenten:**

1. **jotti-Server (Cloud):** Speichert Bestellungen und stellt sie für den Bondruck bereit.
2. **Print-Relay (vor Ort):** Ein kleines Programm, das auf einem Rechner im Festzelt läuft (z.B. Raspberry Pi oder Laptop). Es holt neue Bestellungen vom Server ab und schickt sie an die Drucker.

Das Print-Relay verbindet sich nach außen zum jotti-Server — es muss kein VPN oder Portweiterleitung eingerichtet werden. Solange der Rechner eine Internetverbindung hat und die Drucker im selben lokalen Netzwerk erreichbar sind, funktioniert der Bondruck.

---

## Was steht auf dem Bon?

Jeder Bon enthält die wesentlichen Informationen für die Ausgabestation:

- **Tischnummer** — groß und deutlich, sofort erkennbar beim Aufhängen
- **Position** — was und wie viel (z.B. „3x Pommes (groß)")
- **Kommentar** — Sonderwünsche, falls vorhanden (z.B. „ohne Ketchup")
- **Uhrzeit und Servicekraft** — zur Nachvollziehbarkeit

Preise werden auf dem Bon bewusst nicht angezeigt — die Küche braucht Arbeitsaufträge, keine Preise.

---

## Bonmodus

Der Admin kann pro Kategorie zwischen zwei Modi wählen:

| Modus | Beschreibung | Empfohlen für |
|---|---|---|
| **Pro Position** (Standard) | Jede Position erzeugt einen eigenen Bon | Küche mit mehreren Köchen — jeder Bon ist ein Arbeitsauftrag |
| **Pro Bestellung** | Alle Positionen einer Kategorie auf einem Sammelbon | Kleinere Teams, die lieber einen Überblick pro Bestellung haben |

**Beispiel:** Eine Bestellung mit „3x Pommes, 2x Hefeweizen, 1x Bratwurst":

- **Pro Position:** Küche bekommt 2 Bons (Pommes + Bratwurst), Theke bekommt 1 Bon (Hefeweizen)
- **Pro Bestellung:** Küche bekommt 1 Sammelbon (Pommes + Bratwurst), Theke bekommt 1 Bon (Hefeweizen)

---

## Voraussetzungen

### Drucker

- **ESC/POS-Bondrucker mit Ethernet-Anschluss** (z.B. MUNBYN ITPP047P-UE)
- Anschluss per LAN-Kabel an den Router/Switch des Vereinsfests
- Statische IP-Adresse empfohlen (per DHCP-Reservierung am Router)

> **Tipp:** Die IP-Adresse des Druckers kann über den Selbsttest ermittelt werden: Drucker ausschalten → FEED-Taste gedrückt halten → einschalten → Piepton abwarten → Taste loslassen. Der Drucker druckt seine aktuelle IP-Konfiguration.

### Print-Relay

- Ein Rechner im lokalen Netzwerk des Vereinsfests (z.B. Raspberry Pi, Laptop, Windows-PC)
- Internetverbindung (für die Kommunikation mit dem jotti-Server)
- Im selben Netzwerk wie die Drucker

Das Print-Relay ist ein einzelnes Programm ohne Installation — einfach herunterladen und starten.

---

## Einrichtung

### 1. Drucker konfigurieren (Admin)

Im Admin-Bereich unter **Druckerkonfiguration**:

1. IP-Adresse des Küchenstationsdruckers bei „Essen" eintragen
2. IP-Adresse des Thekendruckers bei „Getränke" eintragen
3. Bonmodus wählen (Pro Position oder Pro Bestellung)
4. Speichern

Kategorien ohne IP-Adresse erzeugen keine Bons (z.B. „Sonstiges" ohne Drucker = kein Druck).

### 2. Print-Relay starten (vor Ort)

```bash
./jotti-relay \
  --backend="https://jotti.meinverein.de" \
  --token="<RELAY_AUTH_TOKEN>" \
  --poll=2
```

| Parameter | Beschreibung | Standard |
|---|---|---|
| `--backend` | URL des jotti-Servers | (erforderlich) |
| `--token` | Authentifizierungs-Token (aus der `.env`-Datei des Servers) | (erforderlich) |
| `--poll` | Abfrageintervall in Sekunden | 2 |
| `--state` | Pfad zur lokalen State-Datei | `relay_state.json` |

Das Relay loggt jeden gedruckten Bon in die Konsole. Wenn kein Drucker erreichbar ist, wartet es und versucht es automatisch erneut.

### 3. Testen

1. Bestellung im Service-Bereich aufnehmen
2. Innerhalb von wenigen Sekunden sollte der Bon gedruckt werden
3. Im Relay-Log erscheint eine Bestätigungsmeldung

---

## Häufige Fragen

### Was passiert, wenn ein Drucker nicht erreichbar ist?

Das Relay versucht bis zu 5 Minuten lang, den Drucker zu erreichen. Danach wird der Auftrag übersprungen und im Log vermerkt. Bereits gedruckte Bons anderer Drucker sind davon nicht betroffen.

### Werden Bons doppelt gedruckt, wenn das Relay neugestartet wird?

Nein. Das Relay merkt sich, welche Bestellungen bereits gedruckt wurden. Bei einem Neustart wird dort weitergemacht, wo aufgehört wurde. In sehr seltenen Fällen (Stromausfall exakt während des Druckens) kann maximal ein einzelner Bon doppelt erscheinen — erkennbar am identischen Zeitstempel.

### Kann ich den Bondruck deaktivieren?

Ja. Einfach keine IP-Adressen in der Druckerkonfiguration eintragen oder das Print-Relay nicht starten. Die Bestellungen funktionieren unabhängig vom Bondruck.

### Welche Drucker werden unterstützt?

Alle ESC/POS-kompatiblen Bondrucker mit Ethernet-Anschluss (TCP Port 9100). Die Architektur ist auf 80mm-Thermodrucker optimiert (48 Zeichen pro Zeile). Getestet mit MUNBYN ITPP047P-UE.

### Funktioniert Bondruck ohne Internetverbindung?

Nein. Das Print-Relay muss den jotti-Server erreichen können, um neue Bestellungen abzuholen. Bei einem Internetausfall werden die Bons nachgedruckt, sobald die Verbindung wiederhergestellt ist.
