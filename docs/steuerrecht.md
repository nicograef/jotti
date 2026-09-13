---
title: Steuerrecht Gastronomie
description: 'Umsatzsteuerrecht der Gastronomie ab 2026: Steuersätze für Speisen und Getränke, Ausnahmen, Kombi-Angebote und Belegpflichtangaben.'
---

Mit dem Steueränderungsgesetz 2025 (Bundesrat-Zustimmung 19. Dezember 2025) wurde § 12 Abs. 2 Nr. 15 UStG zeitlich unbefristet eingeführt. Seit dem 1. Januar 2026 gilt: Speisen 7 %, Getränke 19 %.

## 1. Die zwei Kernsteuersätze

### 1.1 Ermäßigter Steuersatz (7 %): Speisen

Der ermäßigte Satz gilt für alle Abgaben von Speisen im Rahmen einer Restaurations- oder Verpflegungsdienstleistung:

- Warme und kalte Speisen (kein Unterschied)
- Zubereitete Gerichte, belegte Brötchen, Brezeln, abgepackte Lebensmittel (Erdnüsse, Chips)
- Luxusprodukte (Kaviar, Hummer, Austern)
- Alle zugehörigen Serviceleistungen (Servieren, Geschirr bereitstellen, Abspülen)

**Geltungsbereich:** Restaurants, Cafés, Imbisse, Foodtrucks, Kantinen, Event-Caterer, Lieferdienste, unabhängig davon, ob vor Ort verzehrt oder mitgenommen wird.

### 1.2 Regelsteuersatz (19 %): Getränke

Der Regelsteuersatz gilt für die Abgabe von Getränken, ebenfalls unabhängig von der Verzehrsituation:

- Alkoholische Getränke (Bier, Wein, Spirituosen)
- Alkoholfreie Getränke: Softdrinks, Säfte, Mineralwasser
- Kaffee- und Teegetränke auf Wasserbasis

## 2. Ausnahmen und Abgrenzungen

| Produkt / Sachverhalt                                | Steuersatz         | Begründung                                                                     |
| ---------------------------------------------------- | ------------------ | ------------------------------------------------------------------------------ |
| Alle Speisen (vor Ort & To-Go)                       | 7 %                | § 12 Abs. 2 Nr. 15 UStG                                                        |
| Standard-Getränke (Kaffee, Softdrinks, Alkohol)      | 19 %               | Regelsteuersatz                                                                |
| Leitungswasser                                       | 7 %                | Gilt als Lieferung von Trinkwasser (nicht als Getränk im gastronomischen Sinn) |
| Reine Kuhmilch                                       | 7 %                | Grundnahrungsmittel (Anlage 2 zum UStG)                                        |
| Milchmixgetränke (z. B. Cappuccino, Latte Macchiato) | 7 % / 19 %         | 7 % nur bei ≥ 75 % Kuhmilch-Anteil; sonst 19 %                                 |
| Vegane Milchalternativen (Hafer, Soja etc.)          | 19 %               | Gelten rechtlich nicht als Milch, nie begünstigt                               |
| Smoothies (püriertes Obst)                           | 7 %                | Gelten als Speise (feste Nahrung püriert)                                      |
| Fruchtsäfte                                          | 19 %               | Gelten als Getränk                                                             |
| Verkauf im Zweckbetrieb (§ 67a AO)                   | ggf. 0 % / befreit | Ein Zweckbetrieb kann steuerbegünstigt sein                                    |
| Verein als Kleinunternehmer (§ 19 UStG)              | 0 % / befreit      | Vereine mit geringen Umsätzen können von der USt-Pflicht befreit sein          |

## 3. Kombinationsangebote, Menüs und Buffets

Wird ein Pauschalpreis für Speisen und Getränke gemeinsam abgerechnet (Menü, Buffet inkl. Getränke, Spar-Menü mit Softdrink), muss der Gesamtbetrag zwingend in einen Speisen-Anteil (7 %) und einen Getränke-Anteil (19 %) aufgeteilt werden.

### 3.1 Zulässige Aufteilungsmethoden

**Methode A (Kalkulatorische Aufteilung):** Aufteilung nach dem Verhältnis der regulären Einzelverkaufspreise der enthaltenen Komponenten.

**Methode B: 30/70-Pauschalierung** (Vereinfachungsregelung nach Abschn. 10.1 Abs. 12 UStAE; im jotti-Admin heißt dieser Steuersatz „Kombi (70/30)"):

- 30 % des Brutto-Gesamtpreises → Getränke-Anteil → 19 %
- 70 % des Brutto-Gesamtpreises → Speisen-Anteil → 7 %

### 3.2 Einschränkungen

- Kostenbasierte Aufteilung ist unzulässig (Aufteilung nach betrieblichen Kosten wird vom BMF explizit ausgeschlossen).
- **Spar-Menü-Deckelung:** Bei rabattierten Warenzusammenstellungen darf kein anteiliger Einzelpreis den regulären Einzelverkaufspreis übersteigen.

### 3.3 Rechenbeispiel (Methode B)

Menü-Pauschalpreis: 15,00 € brutto

| Anteil          |  Brutto |   Netto | USt-Betrag | Satz |
| --------------- | ------: | ------: | ---------: | ---- |
| Speisen (70 %)  | 10,50 € |  9,81 € |     0,69 € | 7 %  |
| Getränke (30 %) |  4,50 € |  3,78 € |     0,72 € | 19 % |
| Gesamt          | 15,00 € | 13,59 € |     1,41 € |      |

Formel: Netto = Brutto / (1 + Satz); USt = Brutto − Netto.

## 4. Belegausweis und Pflichtangaben

Gemäß § 14 UStG muss jeder Kassenbeleg pro Position ein Steuerkennzeichen (z. B. `A` für 19 %, `B` für 7 %) und im Belegfuß eine Steuermatrix (Netto, Steuerbetrag und Brutto je Steuersatz) ausweisen, keine unaufgeteilte Gesamtsumme ohne Steueraufschlüsselung. Die vollständigen Belegangaben (inkl. TSE-Pflichtfelder) und die technische Belegstruktur: → [compliance.md §5.2](compliance.md#52-pflichtangaben-auf-dem-beleg).
