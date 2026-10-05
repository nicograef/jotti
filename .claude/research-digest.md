# jotti research digest

## Project

jotti is a free, self-hosted point-of-sale system for German non-profits at club festivals and other catering events.
Volunteers take orders, collect payment and cancel per table in the browser of their own phones.
It is an electronic recording system under German cash-register law. It signs bookings through a fiskaly cloud TSE and exports DSFinV-K.

What matters now: the Kassenpflicht bill and its receipt duties, the next DSFinV-K version, and TSE prices and uptime.
Next come installed web apps on iOS and Android, receipt printing, and keeping self-hosting simple for volunteers.

## Lanes

### Cash-register law

Covers: § 146a AO and the AEAO, KassenSichV, GoBD letters, DSFinV-K versions and the ELSTER Kassenmeldung. Also the Kassenpflicht bill with its Belegbereitstellungspflicht, VAT on catering, and non-profit tax law for Vereinsfeste.
Feeds: `https://www.bundesfinanzministerium.de/SiteGlobals/Functions/RSSFeed/DE/Steuern/RSSSteuern.xml`, `https://www.bundesfinanzministerium.de/SiteGlobals/Functions/RSSFeed/DE/Pressemitteilungen/RSSPressemitteilungen.xml`, `https://recht.bund.de/rss/feeds/rss_bgbl-1.xml`, the `Stand` line of `https://www.gesetze-im-internet.de/ao_1977/`, `https://www.gesetze-im-internet.de/kassensichv/` and `https://www.gesetze-im-internet.de/ustg_1980/`, `https://www.bzst.de/DE/Unternehmen/Aussenpruefungen/DigitaleSchnittstelleFinV/digitaleschnittstellefinv_node.html`, Drucksachen on `https://www.bundesrat.de`, `https://www.lobbyregister.bundestag.de` for drafts such as DSFinV-K 3.0, `https://dfka.net/feed/` as a lead.
Queries: `Kassenpflicht Gesetzentwurf`; `Kassenpflicht Verein`; `Bonpflicht OR Belegausgabepflicht OR Belegpflicht`; `Bonpflicht 2028`; `"DSFinV-K" OR KassenSichV OR Kassensicherungsverordnung OR "§ 146a AO"`; `Belegbereitstellungspflicht`; `"§ 146b AO"`; `Kassengesetz`; `Kartenzahlungspflicht`; `Verein (Freigrenze OR Ehrenamtspauschale OR Übungsleiterpauschale OR Vereinsfest)`. Avoid `Registrierkassenpflicht`: it returns Austrian hits.

### fiskaly and TSEs

Covers: fiskaly SIGN DE changes, outages and prices; competing cloud TSEs and prices; BSI TR-03153 versions and the certified-TSE list.
Feeds: `https://status.fiskaly.com/history.rss`, `https://support.fiskaly.com/api/v2/help_center/articles.json?sort_by=updated_at`, `https://www.fiskaly.com/blog`, `https://www.bsi.bund.de/DE/Themen/Unternehmen-und-Organisationen/Standards-und-Zertifizierung/Technische-Richtlinien/TR-nach-Thema-sortiert/tr03153/tr03153_node.html`, TSE price pages of vendors and resellers.
Queries: `"SIGN DE"`, `fiskaly`, `Cloud-TSE`, `Kassensystem TSE`; GitHub search for `fiskaly` and `kassensichv`.

### POS market

Covers: club and event POS products and prices, open-source POS with fiskaly, and club-sector news on cash registers at festivals.
Feeds: `https://news.google.com/rss/search?q=<query>%20when%3A30d&hl=de&gl=DE&ceid=DE:de`; GitHub search for `kassensystem`, `fiskaly` and `kassensichv`.
Queries: `Kassensystem Verein`, `Kassensystem Vereinsfest`, `Kassensystem TSE`, `Kassenpflicht Verein`, `Bonpflicht 2028`.

### Engineering

Covers: installed web apps on iOS and Android, Go, event sourcing on Postgres, Postgres upgrades, ESC/POS network printing. Also JWT and Argon2id guidance, Caddy, acme-dns and Let's Encrypt, and Windows code signing.
Feeds: `https://go.dev/doc/devel/release`, `https://www.postgresql.org/news.rss`, `https://golangweekly.com/rss`, `https://react.statuscode.com/rss`, `https://postgresweekly.com/rss`, `https://lobste.rs/t/go.rss`, `https://webkit.org/feed/`, `https://developer.apple.com/tutorials/data/documentation/safari-release-notes.json`, `https://chromestatus.com/api/v0/features?q=<query>`, `https://github.com/OWASP/CheatSheetSeries/commits/master/cheatsheets/Password_Storage_Cheat_Sheet.md.atom`, `https://letsencrypt.org/feed.xml`.
Queries: `"Go 1.<next>"`, `"Safari <next>"`, `"Web Install API"`, `PWA iOS`, `event sourcing Postgres`, `ESC/POS`, `Argon2id`, `"Let's Encrypt" certificate lifetime`, `DNS-01`, `"Artifact Signing"`, `"SmartScreen reputation"`; GitHub search for `escpos`.

## Reports

`~/Documents/research-digest/jotti/`
