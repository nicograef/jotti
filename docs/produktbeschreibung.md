---
title: Produktbeschreibung
description: 'Abgrenzung von jotti: bewusst ausgeschlossene Funktionen, Architekturprinzip und Einsatzprofil.'
---

jotti ist ein kostenloses Gastronomie-Kassensystem mit einsehbarem Quellcode (Source-Available) für Vereinsfeste, Weihnachtsmärkte, Konzerte und andere Non-Profit-Veranstaltungen. Servicekräfte nehmen Bestellungen direkt auf ihrem Smartphone auf, kassieren und stornieren, alles pro Tisch, alles im Browser.

## 6.2 Was jotti bewusst NICHT ist

jotti ist kein Allzweck-Kassensystem. Folgende Features sind bewusst nicht enthalten: Kartenzahlung/Zahlungsgateway, Reservierungssystem, Inventory/Warenwirtschaft, Lieferservice-Integration, Multi-Standort-Verwaltung, Kundenverwaltung/CRM und Selbstbedienungs-Kiosk (Gäste bestellen/zahlen selbst); der personalbediente Direktverkauf an der Theke ist hingegen enthalten.

Diese bewusste Reduktion ist ein Feature, kein Mangel: Jedes zusätzliche Feature erhöht Komplexität, Wartungsaufwand und Einarbeitungszeit, alles, was ein ehrenamtliches Team bei einem Vereinsfest nicht braucht.

## 6.3 Fiskalkonformität

**Architekturprinzip:** Die Smartphones der Servicekräfte sind reine Eingabegeräte; TSE-Anbindung, Protokollierung und DSFinV-K-Persistenz laufen zentral im Backend, jeder Vorgang ist ein synchroner Backend-Request ohne Offline-Erfassung. Einordnung, Betreiberpflichten und rechtliche Grundlagen: [compliance.md § 2.2](compliance.md#22-kassensicherungsverordnung-kassensichv).

## 6.4 Einsatzprofil

jotti ist nicht geeignet für:

- Dauerbetrieb (Restaurant, Café)
- Kommerzielle Gastro-Betriebe
- Kartenzahlung, NFC, Online-Zahlungen
- Professionelles Gastro-Personal
- Großveranstaltungen mit 100+ Tischen
- Veranstaltungsorte ohne WLAN
