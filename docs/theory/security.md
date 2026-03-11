# Security & Authentifizierung — Theorie

Dieses Dokument ist ein theoretisches Nachschlagewerk für Web-Application-Security. Es erklärt Authentifizierungs- und Autorisierungspatterns, Passwort-Sicherheit, die OWASP Top 10, API- und Frontend-Security, Secrets Management, TLS und Security Testing.

---

## Inhaltsverzeichnis

1. [Warum Security-Architektur?](#1-warum-security-architektur)
2. [Authentication-Patterns](#2-authentication-patterns)
3. [Authorization-Modelle](#3-authorization-modelle)
4. [Passwort-Sicherheit](#4-passwort-sicherheit)
5. [OWASP Top 10 (2021)](#5-owasp-top-10-2021)
6. [API-Security](#6-api-security)
7. [Frontend-Security](#7-frontend-security)
8. [Secrets Management](#8-secrets-management)
9. [TLS & Certificate Management](#9-tls--certificate-management)
10. [Security Testing](#10-security-testing)
11. [Entscheidungsmatrix: Auth-Pattern nach Anwendungsfall](#11-entscheidungsmatrix-auth-pattern-nach-anwendungsfall)
12. [Referenzen](#12-referenzen)

---

## 1. Warum Security-Architektur?

Security ist kein Feature, das man am Ende hinzufügt — sie ist ein **Querschnittsthema**, das jede Schicht einer Anwendung betrifft. Ein Architekturansatz ohne Sicherheitsbetrachtung ist unvollständig.

### 1.1 Security als Architekturprinzip

Das klassische Modell betrachtet Security als „Afterthought" — man baut die Funktion, dann sichert man sie ab. Moderne Ansätze (DevSecOps, Shift-Left Security) integrieren Security von Anfang an:

- **Design-Phase:** Threat Modeling, Security Requirements
- **Entwicklung:** Secure Coding Guidelines, Code Reviews, SAST
- **CI/CD:** Dependency Scanning, DAST, Secret Scanning
- **Betrieb:** Monitoring, Alerting, Incident Response, Penetration Testing

**Security-Schichten (Defense in Depth):**

```
┌──────────────────────────────────────────┐
│  Anwendungsschicht (Input Validation,    │
│  Auth, AuthZ, Business Logic Security)   │
├──────────────────────────────────────────┤
│  API-Schicht (Rate Limiting, CORS,       │
│  Security Headers, TLS)                  │
├──────────────────────────────────────────┤
│  Infrastruktur (Firewall, VPN,           │
│  Netzwerk-Segmentierung)                 │
├──────────────────────────────────────────┤
│  Betriebssystem & Container              │
│  (Updates, CIS Benchmarks, Rootless)     │
└──────────────────────────────────────────┘
```

### 1.2 Threat Modeling (STRIDE)

Threat Modeling ist der strukturierte Prozess, potenzielle Bedrohungen frühzeitig zu identifizieren. Das **STRIDE-Modell** kategorisiert Bedrohungen nach sechs Typen:

| Buchstabe | Bedrohung              | Beschreibung                                   | Gegenmittel                             |
| --------- | ---------------------- | ---------------------------------------------- | --------------------------------------- |
| **S**     | Spoofing               | Vortäuschen einer anderen Identität            | Authentifizierung, MFA                  |
| **T**     | Tampering              | Manipulation von Daten in Transit oder at Rest | Integritätsprüfung, HMAC, TLS           |
| **R**     | Repudiation            | Abstreiten einer Handlung ohne Nachweis        | Audit Logging, digitale Signaturen      |
| **I**     | Information Disclosure | Unberechtigter Zugriff auf vertrauliche Daten  | Verschlüsselung, Access Control         |
| **D**     | Denial of Service      | Störung der Verfügbarkeit                      | Rate Limiting, WAF, Auto-Scaling        |
| **E**     | Elevation of Privilege | Unberechtiger Zugewinn höherer Rechte          | Least Privilege, RBAC, Input Validation |

**STRIDE-Prozess:**

1. Systemarchitektur als Datenflussdiagramm (DFD) modellieren
2. Vertrauensgrenzen einzeichnen (wo wechselt der Kontext?)
3. Für jede Komponente alle sechs STRIDE-Kategorien durchgehen
4. Risiko bewerten (Wahrscheinlichkeit × Schadenspotenzial)
5. Gegenmittel definieren und in Design einarbeiten

---

## 2. Authentication-Patterns

Authentifizierung beantwortet die Frage: **Wer bist du?** Sie ist der Einstiegspunkt jeder gesicherten Anwendung.

### 2.1 Session-basierte Authentifizierung

Der klassische Web-Ansatz: Nach dem Login erzeugt der Server eine Session und gibt dem Client eine **Session-ID** (meist als Cookie).

**Ablauf:**

```
Client                    Server
  │  POST /login            │
  │ ─────────────────────►  │
  │                         │ Session erstellen
  │                         │ Session-ID speichern (DB/Cache)
  │  Set-Cookie: sid=abc123 │
  │ ◄─────────────────────  │
  │                         │
  │  GET /dashboard         │
  │  Cookie: sid=abc123     │
  │ ─────────────────────►  │
  │                         │ Session-ID validieren
  │  200 OK                 │
  │ ◄─────────────────────  │
```

**Charakteristika:**

| Aspekt          | Beschreibung                                               |
| --------------- | ---------------------------------------------------------- |
| **Stateful**    | Server speichert Session-State (DB, Redis, In-Memory)      |
| **Revozierbar** | Session kann serverseitig sofort ungültig gemacht werden   |
| **Skalierung**  | Sticky Sessions oder zentraler Session-Store (Redis) nötig |
| **Cookies**     | httpOnly + Secure + SameSite=Strict empfohlen              |

**Schwächen:** Session-Hijacking (wenn Cookie gestohlen), CSRF (Cookie wird automatisch mitgesendet), horizontale Skalierung erfordert Session-Sharing.

### 2.2 Token-basierte Authentifizierung (JWT)

JSON Web Tokens (RFC 7519) sind selbst-beschreibende, signierte Token. Der Server speichert **keinen State** — alle nötigen Informationen sind im Token kodiert.

**Struktur:**

```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9   ← Header (Base64url)
.eyJzdWIiOiIxMjMiLCJyb2xlIjoiYWRtaW4ifQ   ← Payload (Base64url)
.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c  ← Signatur
```

**Header:** Algorithmus + Typ  
**Payload (Claims):** `sub` (Subject/UserID), `role`, `iat` (Issued At), `exp` (Expiration)  
**Signatur:** HMAC oder RSA-Signatur über Header + Payload

**HS256 vs. RS256:**

| Algorithmus | Signierung           | Verifikation  | Wann sinnvoll                             |
| ----------- | -------------------- | ------------- | ----------------------------------------- |
| **HS256**   | Shared Secret (HMAC) | Shared Secret | Monolith, alle Services kennen das Secret |
| **RS256**   | Private Key (RSA)    | Public Key    | Microservices, externe Token-Validierung  |
| **ES256**   | Private Key (ECDSA)  | Public Key    | Wie RS256, kleinere Schlüssel             |

**Access + Refresh Token Pattern:**

```
┌─────────┐  Kurzlebig (5-15 min)   ┌─────────────┐
│ Access  │ ──────────────────────► │ API-Aufrufe  │
│ Token   │                         │              │
└─────────┘                         └─────────────┘

┌─────────┐  Langlebig (7-30 Tage)  ┌──────────────┐
│ Refresh │ ──────────────────────► │ Neues Access  │
│ Token   │                         │ Token holen   │
└─────────┘                         └──────────────┘
```

**Token Rotation:** Bei jeder Refresh-Token-Nutzung wird ein neues Refresh-Token ausgestellt. Das alte wird invalidiert. Gestohlene Refresh-Tokens werden so schnell erkannt.

**Bekannte Sicherheitslücken:**

- **`alg: none`-Angriff:** Angreifer entfernt die Signatur und setzt `alg` auf `none`. Fix: Immer den erwarteten Algorithmus explizit prüfen.
- **Key Confusion (RS256 → HS256):** Angreifer signiert mit dem Public Key als HMAC-Secret. Fix: Algorithmus serverseitig fest vorgeben.
- **Kurze Secrets:** HMAC-Secrets müssen mindestens 256 Bit stark sein.

### 2.3 OAuth2 / OpenID Connect (OIDC)

OAuth2 ist ein **Autorisierungs**-Framework — es ermöglicht einer Anwendung, im Namen eines Nutzers auf Ressourcen zuzugreifen. OpenID Connect (OIDC) fügt eine **Authentifizierungsschicht** hinzu.

**Authorization Code Flow mit PKCE (empfohlen für SPAs und mobile Apps):**

```
Browser          Auth-Server         Resource-Server
   │                   │                    │
   │  1. Redirect mit  │                    │
   │     code_challenge│                    │
   │ ─────────────────►│                    │
   │                   │ Login + Consent     │
   │  2. Auth-Code     │                    │
   │ ◄─────────────────│                    │
   │                   │                    │
   │  3. Auth-Code +   │                    │
   │     code_verifier │                    │
   │ ─────────────────►│                    │
   │                   │ PKCE prüfen        │
   │  4. Access Token  │                    │
   │     + ID Token    │                    │
   │ ◄─────────────────│                    │
   │                                        │
   │  5. API-Request mit Access Token       │
   │ ──────────────────────────────────────►│
```

**PKCE (Proof Key for Code Exchange):** Verhindert Authorization-Code-Hijacking. Der Client erzeugt einen zufälligen `code_verifier`, sendet dessen Hash (`code_challenge`) beim ersten Request und beweist später den Besitz mit dem Original-Verifier.

**Tokens in OIDC:**

| Token             | Format     | Inhalt                             | Verwendung                  |
| ----------------- | ---------- | ---------------------------------- | --------------------------- |
| **Access Token**  | JWT/Opaque | Berechtigungen (Scopes)            | API-Zugriff                 |
| **ID Token**      | JWT        | Nutzeridentität (sub, email, name) | Authentifizierungs-Nachweis |
| **Refresh Token** | Opaque     | Token-Erneuerung                   | Neues Access Token holen    |

### 2.4 API Keys

API Keys sind statische Credentials für **Maschine-zu-Maschine**-Kommunikation (M2M). Kein interaktiver Login, kein Session-State.

**Best Practices:**

- Ausreichend lange, zufällige Keys (256+ Bit Entropie)
- Keys hashen in der Datenbank (wie Passwörter — SHA-256 reicht hier, da keine Brute-Force-Gefahr bei langen Keys)
- Scoping: Jeder Key hat genau definierte Berechtigungen
- Rotation: Mechanismus für Key-Rotation ohne Downtime
- Rate Limiting pro Key

### 2.5 Passkeys / WebAuthn

**WebAuthn** (Web Authentication API) ist ein W3C-Standard für passwortlose Authentifizierung. Nutzer authentifizieren sich mit einem Authenticator (Gerät, Biometrie, Security Key).

**Funktionsprinzip:**

```
Registrierung:                         Authentifizierung:
1. Server sendet Challenge             1. Server sendet Challenge
2. Authenticator signiert mit          2. Authenticator signiert mit
   Private Key (lokal gespeichert)        Private Key
3. Public Key + Attestation an Server  3. Signatur an Server
4. Server speichert Public Key         4. Server verifiziert mit Public Key
```

**Vorteile:** Phishing-resistent (Credential ist domaingebunden), keine Passwörter, biometrische Verifikation lokal auf dem Gerät.

---

## 3. Authorization-Modelle

Autorisierung beantwortet die Frage: **Was darfst du tun?** Sie folgt der Authentifizierung und prüft Berechtigungen für konkrete Aktionen.

### 3.1 RBAC — Role-Based Access Control

Berechtigungen werden **Rollen** zugewiesen, Rollen werden Nutzern zugewiesen. Einfach, gut verständlich, weit verbreitet.

```
Nutzer ──► Rolle ──► Berechtigung
Max     ──► Admin ──► users:create, users:delete, products:write
Lisa    ──► Service ──► orders:create, orders:read
```

**Rollenhierarchien:** Admin erbt alle Berechtigungen von Service. Reduziert Redundanz.

**Wann RBAC:** Klar definierte Rollen, wenige Dimensionen, keine feingranulare Kontext-Abhängigkeit nötig.

**Schwächen:** Role Explosion (zu viele Rollen für Spezialfälle), keine Kontext-Abhängigkeit (z.B. „nur eigene Ressourcen").

### 3.2 ABAC — Attribute-Based Access Control

Entscheidungen basieren auf **Attributen** von Subjekt (Nutzer), Objekt (Ressource) und Umgebung (Zeit, IP, etc.).

```
Policy: ALLOW IF
  subject.role == "manager" AND
  resource.department == subject.department AND
  environment.time BETWEEN 08:00 AND 18:00
```

**Wann ABAC:** Komplexe, kontextsensitive Zugriffsregeln; feingranulare Steuerung; regulierte Umgebungen (GDPR, HIPAA).

**Schwächen:** Komplex zu implementieren und zu debuggen; Policy-Konflikte schwer erkennbar; Performance bei vielen Attributen.

### 3.3 ReBAC — Relationship-Based Access Control

Berechtigungen folgen aus **Beziehungen** zwischen Entitäten. Grundlage: Google Zanzibar (2019) — das globale Autorisierungssystem von Google, das für Milliarden Objekte skaliert.

**Zanzibar-Modell:**

```
Relation Tuples (Fakten):
  document:readme#owner@alice       ← Alice ist Owner von readme
  document:readme#viewer@bob        ← Bob ist Viewer
  document:readme#viewer@org:eng#member  ← Alle Mitglieder von org:eng sind Viewer

Check (Anfrage):
  "Darf alice document:readme editieren?" → YES (alice ist owner)
  "Darf charlie document:readme lesen?"  → NO (keine Relation definiert)
```

**Open-Source-Implementierungen:**

| Tool         | Beschreibung                         | Verwendungskontext                   |
| ------------ | ------------------------------------ | ------------------------------------ |
| **OpenFGA**  | Open-Source Zanzibar, von Auth0/Okta | Google Docs-ähnliche Sharing-Modelle |
| **SpiceDB**  | Zanzibar-Implementierung von Authzed | Konsistente, globale Autorisierung   |
| **Ory Keto** | Permission Service nach Zanzibar     | Microservices-Autorisierung          |

**Wann ReBAC:** Kollaborationsplattformen, Dokument-Sharing, Multi-Tenant-Systeme, komplexe Ressourcen-Hierarchien.

### 3.4 Policy Engines

Policy Engines entkoppeln die Autorisierungslogik vom Anwendungscode.

**OPA (Open Policy Agent):**

```rego
# policy.rego
package authz

default allow = false

allow {
    input.method == "GET"
    input.path == ["api", "products"]
    input.user.role == "service"
}

allow {
    input.user.role == "admin"
}
```

OPA empfängt Anfragen als JSON, wertet Policies aus (Rego-Sprache) und gibt `allow: true/false` zurück. Einsatzgebiet: Kubernetes (OPA Gatekeeper), API Gateways, Microservices.

**Cedar (AWS):** Stark typisierte Policy-Sprache, formal verifizierbar. Eingesetzt in Amazon Verified Permissions.

---

## 4. Passwort-Sicherheit

### 4.1 Hashing vs. Verschlüsselung

Passwörter müssen **gehasht**, nicht verschlüsselt werden. Hashing ist eine Einwegfunktion — aus dem Hash kann das Klartextpasswort nicht wiederhergestellt werden. Verschlüsselung ist eine Zweiwegfunktion — bei Kompromittierung des Keys sind alle Passwörter im Klartext.

### 4.2 Hashing-Algorithmen im Vergleich

**Empfehlung (OWASP Password Storage Cheat Sheet):**

| Priorität | Algorithmus  | Minimale Parameter                                    | Wann wählen                                    |
| --------- | ------------ | ----------------------------------------------------- | ---------------------------------------------- |
| ✅ 1.     | **Argon2id** | 19 MiB Memory, 2 Iterationen, 1 Parallelismus-Grad    | Standard-Empfehlung (Memory-Hard)              |
| ✅ 2.     | **scrypt**   | CPU/Memory-Cost ≥ 2^17, Block-Size 8, Parallelismus 1 | Wenn Argon2id nicht verfügbar                  |
| ✅ 3.     | **bcrypt**   | Work Factor ≥ 10, Passwort-Limit 72 Bytes beachten    | Legacy-Systeme, breite Framework-Unterstützung |
| ✅ 4.     | **PBKDF2**   | HMAC-SHA-256, ≥ 600.000 Iterationen                   | FIPS-140-Compliance erforderlich               |
| ❌        | MD5, SHA-1   | —                                                     | Nie für Passwörter — zu schnell                |
| ❌        | SHA-256/512  | —                                                     | Nie für Passwörter ohne Work Factor            |

**Argon2-Varianten:**

| Variante     | Schutz gegen          | Empfehlung                                |
| ------------ | --------------------- | ----------------------------------------- |
| Argon2d      | GPU-Angriffe          | Für Kryptowährungen                       |
| Argon2i      | Side-Channel-Angriffe | Für Passwort-Hashing in einigen Kontexten |
| **Argon2id** | Beide                 | **Standard-Empfehlung**                   |

### 4.3 Salting & Peppering

**Salt:** Ein einzigartiger, zufälliger Wert, der vor dem Hashing an das Passwort angehängt wird. Schützt gegen Rainbow-Table-Angriffe und verhindert, dass identische Passwörter identische Hashes erzeugen. Moderne Algorithmen (Argon2, bcrypt) verwalten Salts automatisch.

**Pepper:** Ein anwendungsweiter, geheimer Wert, der zusätzlich zum Salt verwendet wird. Im Unterschied zum Salt wird der Pepper **nicht** in der Datenbank gespeichert, sondern als Anwendungs-Secret. Schützt auch bei vollständigem Datenbankdiebstahl.

### 4.4 NIST-Empfehlungen (SP 800-63B, 2024)

Das National Institute of Standards and Technology (NIST) hat seine Passwort-Guidelines 2024 erheblich überarbeitet:

**Was NIST empfiehlt:**

- Mindestlänge: 8 Zeichen (Empfehlung: 15+)
- Maximallänge: Mindestens 64 Zeichen erlauben
- Alle druckbaren ASCII-Zeichen sowie Unicode erlauben
- Neue Passwörter gegen bekannte kompromittierte Passwort-Listen prüfen (z.B. HaveIBeenPwned API)
- MFA (Multi-Factor Authentication) anbieten

**Was NIST nicht mehr empfiehlt:**

- Regelmäßige erzwungene Passwort-Rotation (erhöht Vorhersagbarkeit)
- Komplexitätsregeln (Sonderzeichen, Großbuchstaben — führen zu schwachen Mustern wie `Password1!`)
- Sicherheitsfragen als zweiten Faktor

---

## 5. OWASP Top 10 (2021)

Das Open Web Application Security Project (OWASP) veröffentlicht regelmäßig die zehn kritischsten Sicherheitsrisiken für Web-Anwendungen. Die Versionen von 2021 und 2025 bilden die aktuelle Grundlage.

### A01 — Broken Access Control

**Was:** Nutzer können auf Ressourcen oder Funktionen zugreifen, für die sie keine Berechtigung haben.

**Beispiele:** Direkter Objekt-Referenz-Zugriff (`/api/orders/12345` ohne Prüfung ob Order dem Nutzer gehört), fehlende Funktions-Zugriffskontrolle (Admin-Endpunkte ohne Prüfung der Rolle), CORS-Fehlkonfiguration, Path Traversal.

**Mitigation:** Least-Privilege-Prinzip durchsetzen, serverseitige Zugriffskontrollen auf allen Ebenen (nie nur im Frontend), Besitzer-Prüfung bei Ressourcenzugriff, automatisierte Tests für Zugriffskontrolle.

### A02 — Cryptographic Failures

**Was:** Vertrauliche Daten werden unverschlüsselt übertragen oder gespeichert, oder schwache Kryptographie wird eingesetzt.

**Beispiele:** Passwörter als Klartext in der Datenbank, veraltete Algorithmen (MD5, SHA-1, DES), unverschlüsselte HTTP-Verbindungen, schwache TLS-Konfiguration (TLS 1.0/1.1), fehlende `Secure`-Flag bei Cookies.

**Mitigation:** TLS für alle Verbindungen, starke Hashing-Algorithmen (Argon2id, bcrypt), Verschlüsselung sensitiver Daten at Rest, Datenschutz-Klassifizierung (welche Daten sind schutzbedürftig?).

### A03 — Injection

**Was:** Nicht-vertrauenswürdige Daten werden als Teil eines Befehls oder einer Query interpretiert.

**Typen:** SQL-Injection, NoSQL-Injection, OS-Command-Injection, LDAP-Injection, XSS (HTML-Injection).

**SQL-Injection-Beispiel:**

```sql
-- Verwundbar:
SELECT * FROM users WHERE username = '" + input + "'";
-- Bei input = "admin' OR '1'='1":
SELECT * FROM users WHERE username = 'admin' OR '1'='1'

-- Sicher (Prepared Statements):
SELECT * FROM users WHERE username = $1
```

**Mitigation:** Prepared Statements / Parameterisierte Queries, ORM/Query Builder, Input Validation (Whitelist), Least-Privilege für DB-User, WAF als ergänzende Schicht.

### A04 — Insecure Design

**Was:** Fehlende oder ineffektive Sicherheitsprinzipien im Entwurf der Anwendung — Sicherheitsrisiken die durch besseres Design von Anfang an vermeidbar wären.

**Beispiele:** Fehlende Rate-Limits bei Login (Brute Force), kein Schutz gegen Massenregistrierung, fehlende Trennung sensitiver Workflows.

**Mitigation:** Threat Modeling in der Design-Phase, Secure Design Patterns, Security Stories in agilen Prozessen.

### A05 — Security Misconfiguration

**Was:** Unsichere Standardkonfigurationen, unnötige Features aktiviert, fehlende Sicherheits-Hardening.

**Beispiele:** Default-Credentials nicht geändert, Stack Traces in Produktions-Fehlerseiten, unnötige offene Ports, fehlendes Security-Header-Set.

**Security Headers-Checkliste:**

| Header                            | Zweck                                            |
| --------------------------------- | ------------------------------------------------ |
| `Strict-Transport-Security`       | Erzwingt HTTPS (HSTS)                            |
| `Content-Security-Policy`         | Verhindert XSS, steuert erlaubte Quellen         |
| `X-Content-Type-Options: nosniff` | Verhindert MIME-Sniffing                         |
| `X-Frame-Options: DENY`           | Verhindert Clickjacking                          |
| `Referrer-Policy`                 | Steuert Referrer-Header                          |
| `Permissions-Policy`              | Beschränkt Browser-APIs (Kamera, Mikrofon, etc.) |

### A06 — Vulnerable and Outdated Components

**Was:** Verwendung von Bibliotheken, Frameworks oder Systemen mit bekannten Schwachstellen.

**Mitigation:** Dependency-Scanning in CI/CD (Dependabot, Snyk, Trivy), regelmäßige Updates, SBOM (Software Bill of Materials), Monitoring von CVEs für verwendete Komponenten.

### A07 — Identification and Authentication Failures

**Was:** Schwächen in Authentifizierungsimplementierungen: schwache Passwort-Policies, fehlende Brute-Force-Schutz, unsichere Session-Verwaltung.

**Beispiele:** Session-IDs in URLs, fehlende Session-Invalidierung beim Logout, fehlender MFA-Schutz für privilegierte Konten, Credential Stuffing ohne Schutz.

**Mitigation:** Starkes Passwort-Hashing, Rate Limiting für Login-Endpunkte, MFA für kritische Accounts, sichere Session-Verwaltung (httpOnly, Secure, SameSite Cookies).

### A08 — Software and Data Integrity Failures

**Was:** Code oder Daten werden ohne Integritätsprüfung verwendet. Betrifft CI/CD-Pipelines, Deserialisierung und Software-Updates.

**Beispiele:** Fehlende Signaturprüfung für Software-Updates, unsichere Deserialisierung, Angreifer in der CI/CD-Pipeline (Supply Chain Attacks).

**Mitigation:** Subresource Integrity (SRI) für externe Scripts, Code-Signierung, Verifikation von Abhängigkeiten (Checksums), Dependency Pinning.

### A09 — Security Logging and Monitoring Failures

**Was:** Fehlende oder unzureichende Protokollierung sicherheitsrelevanter Ereignisse — verhindert Erkennung und Reaktion auf Angriffe.

**Was geloggt werden sollte:**

- Fehlgeschlagene Login-Versuche (mit Timestamp, IP)
- Zugriffsverletzungen (403-Fehler)
- Admin-Aktionen (User erstellt/gelöscht, Config geändert)
- Abnormale Muster (viele 404/500 in kurzer Zeit)

**Mitigation:** Strukturiertes Logging (JSON), zentralisiertes Log-Management (ELK, Loki), Alerting auf sicherheitsrelevante Events, Logs vor Manipulation schützen (append-only, extern speichern).

### A10 — Server-Side Request Forgery (SSRF)

**Was:** Die Anwendung lädt eine Ressource von einem URL, der vom Angreifer kontrolliert wird. Der Server macht Requests im Namen des Angreifers — potenziell zu internen Diensten, die von außen nicht erreichbar sind.

**Beispiel:** `POST /api/fetch-image?url=http://169.254.169.254/latest/meta-data/` (AWS Metadata-Service).

**Mitigation:** URL-Whitelist (nur erlaubte Domains), Firewall-Regeln (interne Dienste nicht über HTTP erreichbar), Netzwerk-Segmentierung, Deaktivierung unnötiger URL-Fetch-Features.

---

## 6. API-Security

### 6.1 Input Validation

**Whitelist > Blacklist:** Definiere, was erlaubt ist (Whitelist), nicht was verboten ist (Blacklist). Blacklists sind unvollständig — Angreifer finden immer neue Umgehungen.

**Validierungsebenen:**

```
HTTP-Schicht   → Content-Type, Content-Length
Schema-Ebene   → Datentypen, Felder, Format (Zod/zog)
Business-Ebene → Geschäftsregeln (z.B. Menge > 0, Datum in der Zukunft)
Datenbank      → Constraints (NOT NULL, CHECK, FOREIGN KEY)
```

**Nie vertrauen:** Request-Parameter, Headers, Body, Pfad-Segmente, Query-Parameter, Dateinamen.

### 6.2 CORS — Cross-Origin Resource Sharing

CORS ist ein Browser-Sicherheitsmechanismus basierend auf der **Same-Origin-Policy**: Browser blockieren standardmäßig HTTP-Requests von einer Domain zu einer anderen.

**CORS-Header:**

| Header                             | Bedeutung                                             |
| ---------------------------------- | ----------------------------------------------------- |
| `Access-Control-Allow-Origin`      | Erlaubte Origins (nie `*` für authentifizierte APIs!) |
| `Access-Control-Allow-Methods`     | Erlaubte HTTP-Methoden                                |
| `Access-Control-Allow-Headers`     | Erlaubte Request-Header                               |
| `Access-Control-Allow-Credentials` | Ob Cookies mitgesendet werden dürfen (`true`)         |
| `Access-Control-Max-Age`           | Wie lange Preflight-Ergebnis gecacht werden darf      |

**Sicherheitsregel:** Wenn `Allow-Credentials: true` gesetzt wird, darf `Allow-Origin` **nicht** `*` sein. Mit `*` und Credentials werden alle Cookies und Auth-Headers exponiert.

### 6.3 CSRF — Cross-Site Request Forgery

CSRF nutzt aus, dass Browser Cookies automatisch mitsenden. Eine bösartige Website kann Requests an eine andere Domain triggern, wobei die Cookies des eingeloggten Nutzers mitgesendet werden.

**Schutzmaßnahmen:**

| Methode                  | Beschreibung                                                          | Wann sinnvoll                 |
| ------------------------ | --------------------------------------------------------------------- | ----------------------------- |
| **SameSite Cookies**     | `SameSite=Strict` oder `Lax` — Browser sendet Cookie nicht bei CSRF   | Einfachste, modernste Lösung  |
| **CSRF-Token**           | Server erzeugt zufälliges Token, das im Request-Body mitgesendet wird | Wenn SameSite nicht ausreicht |
| **Double Submit Cookie** | CSRF-Token als Cookie + Request-Header — beide müssen übereinstimmen  | Stateless CSRF-Schutz         |
| **Origin-Header prüfen** | Server prüft `Origin` oder `Referer` Header                           | Ergänzende Maßnahme           |

**Wichtig:** Token-basierte APIs (JWT im `Authorization`-Header, nicht als Cookie) sind von CSRF nicht betroffen, weil Browser den `Authorization`-Header nicht automatisch setzen.

### 6.4 Rate Limiting & Throttling

Rate Limiting schützt vor Brute-Force-Angriffen, Credential Stuffing und DoS.

**Strategien:**

| Strategie          | Beschreibung                                                            |
| ------------------ | ----------------------------------------------------------------------- |
| **Fixed Window**   | X Requests pro Zeitfenster (z.B. 100 req/min)                           |
| **Sliding Window** | Wie Fixed, aber rollierende Berechnung — gleichmäßigerer Schutz         |
| **Token Bucket**   | Token werden kontinuierlich hinzugefügt, jeder Request verbraucht einen |
| **Leaky Bucket**   | Requests werden in gleichmäßigem Tempo verarbeitet                      |

**Granularität:** Nach IP, nach Nutzer-ID, nach API-Key, nach Endpoint-Typ (Login-Endpoints strenger als normale APIs).

---

## 7. Frontend-Security

### 7.1 XSS — Cross-Site Scripting

XSS-Angriffe injizieren ausführbares JavaScript in eine Webseite, das im Browser anderer Nutzer ausgeführt wird.

**Typen:**

| Typ               | Beschreibung                                                       |
| ----------------- | ------------------------------------------------------------------ |
| **Reflected XSS** | Script im Request, direkt in Response gespiegelt                   |
| **Stored XSS**    | Script in Datenbank gespeichert, bei jedem Seitenaufruf ausgeführt |
| **DOM-based XSS** | Manipulation des DOM durch JavaScript ohne Server-Beteiligung      |

**React und XSS:** React escapt Output standardmäßig beim Rendern von JSX. `{userInput}` wird sicher gerendert. Risiko entsteht nur durch `dangerouslySetInnerHTML` — explizit benannt als Warnung.

```tsx
// Sicher — React escapt automatisch:
<div>{userInput}</div>

// Gefährlich — nur für vertrauenswürdigen HTML:
<div dangerouslySetInnerHTML={{ __html: sanitizedHtml }} />
// → Immer DOMPurify o.ä. zur Sanitisierung verwenden
```

**Content Security Policy (CSP):** HTTP-Header, der dem Browser vorschreibt, welche Quellen für Scripts, Styles, Bilder etc. erlaubt sind. Verhindert XSS selbst wenn Code injiziert wird.

```
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
```

### 7.2 Token-Handling: httpOnly Cookie vs. localStorage vs. Memory

Wo Tokens im Browser gespeichert werden, hat erhebliche Security-Implikationen:

| Speicherort         | XSS-Schutz | CSRF-Schutz          | Persistenz  | Empfehlung                                      |
| ------------------- | ---------- | -------------------- | ----------- | ----------------------------------------------- |
| **httpOnly Cookie** | ✅ Ja      | ⚠️ Nein (→ SameSite) | ✅ Ja       | Beste Option mit `SameSite=Strict` + `Secure`   |
| **localStorage**    | ❌ Nein    | ✅ Ja                | ✅ Ja       | Vermeiden für Auth-Tokens                       |
| **sessionStorage**  | ❌ Nein    | ✅ Ja                | Tab-Session | Besser als localStorage, aber noch XSS-anfällig |
| **Memory (JS-Var)** | ✅ Ja      | ✅ Ja                | ❌ Nein     | Sicherste Option, aber kein Page Reload         |

**Empfehlung:** httpOnly Cookie mit `SameSite=Strict`, `Secure` und CSRF-Schutz.

### 7.3 Subresource Integrity (SRI)

SRI ermöglicht, externe Ressourcen (Scripts, Stylesheets von CDNs) durch einen Hash zu verifizieren:

```html
<script
  src="https://cdn.example.com/library.js"
  integrity="sha384-abc123..."
  crossorigin="anonymous"
></script>
```

Der Browser lädt die Ressource nur, wenn der Hash übereinstimmt. Schützt gegen kompromittierte CDNs.

---

## 8. Secrets Management

### 8.1 Environment Variables (12-Factor App)

Das [12-Factor App](https://12factor.net/config)-Prinzip besagt: Konfiguration (Secrets, Connection Strings, API Keys) gehört in **Environment Variables**, nicht in den Code.

**Vorteile:** Keine Secrets im Git-Repository, einfache Rotation, umgebungsspezifische Konfiguration.

**Risiken:** Environment Variables sind nicht verschlüsselt, potenziell im Prozess-Speicher lesbar, in Container-Logs sichtbar.

### 8.2 .env-Dateien

`.env`-Dateien sind praktisch für die Entwicklung, aber:

- **Niemals** in Git committen (`.gitignore` + `.env.example` mit Platzhaltern)
- In Produktion durch echtes Secrets Management ersetzen
- `.env`-Dateien nicht in Container-Images einbacken

### 8.3 Secrets Management Tools

| Tool                                      | Beschreibung                                                 | Wann sinnvoll                          |
| ----------------------------------------- | ------------------------------------------------------------ | -------------------------------------- |
| **HashiCorp Vault**                       | Enterprise Secret Store, Dynamic Secrets, Audit Logs         | Komplexe Infrastruktur, viele Services |
| **SOPS**                                  | Verschlüsselte Secret-Dateien im Git (KMS/PGP/Age)           | GitOps-Workflows, kleinere Teams       |
| **Sealed Secrets**                        | Kubernetes-nativer Ansatz, verschlüsselte Secrets im Cluster | Kubernetes-Deployments                 |
| **AWS Secrets Manager / Parameter Store** | Cloud-nativer Secret Store                                   | AWS-Umgebungen                         |
| **1Password Secrets Automation**          | API-basierter Zugriff auf 1Password Vaults                   | Teams die 1Password bereits nutzen     |

### 8.4 Secret Scanning in CI/CD

Git-Repositories müssen vor dem Commit auf versehentlich eingecheckte Secrets gescannt werden:

- **Pre-commit Hooks:** `git-secrets`, `detect-secrets`, `gitleaks`
- **CI/CD-Integration:** GitHub Secret Scanning (automatisch), `truffleHog`, GitGuardian
- **Rotation-Policy:** Gestohlene Secrets sofort rotieren, auch wenn der Schaden unklar ist

---

## 9. TLS & Certificate Management

### 9.1 TLS-Grundlagen

Transport Layer Security (TLS) verschlüsselt die Verbindung zwischen Client und Server. TLS 1.3 (2018) ist der aktuelle Standard.

**TLS-Handshake (vereinfacht TLS 1.3):**

```
Client                        Server
  │  ClientHello               │
  │  (TLS-Version, Cipher, Random) ────────────────────────►  │
  │                            │
  │  ServerHello               │
  │  (Gewählter Cipher, Random)│
  │  Certificate               │
  │  (Öffentliches Zertifikat) │
  │  CertificateVerify         │
  │  Finished                  │
  │ ◄─────────────────────────  │
  │                            │
  │  Finished ─────────────────►  │
  │                            │
  │ ═══ Verschlüsselte Kommunikation ═══  │
```

**TLS-Versionen:**

| Version | Status        | Anmerkung                                          |
| ------- | ------------- | -------------------------------------------------- |
| TLS 1.0 | ❌ Deprecated | Anfällig für BEAST, POODLE — nicht mehr verwenden  |
| TLS 1.1 | ❌ Deprecated | Ebenfalls veraltet                                 |
| TLS 1.2 | ⚠️ Akzeptabel | Noch weit verbreitet, sichere Cipher-Suites nötig  |
| TLS 1.3 | ✅ Empfohlen  | Schneller Handshake, Forward Secrecy standardmäßig |

### 9.2 Zertifikat-Typen

| Typ                             | Validierungstiefe        | Ausstellzeit    | Wann sinnvoll                                      |
| ------------------------------- | ------------------------ | --------------- | -------------------------------------------------- |
| **DV (Domain Validated)**       | Domainbesitz             | Minuten-Stunden | Meiste Webanwendungen, Entwicklung                 |
| **OV (Organization Validated)** | Organisation             | Tage            | Unternehmens-Websites                              |
| **EV (Extended Validated)**     | Ausführliche Prüfung     | Wochen          | Banken, E-Commerce (Grüner Browser-Balken obsolet) |
| **Wildcard**                    | Domain + alle Subdomains | Minuten (LE)    | Viele Subdomains                                   |

### 9.3 Let's Encrypt & ACME

Let's Encrypt ist eine kostenlose Certificate Authority (CA), die DV-Zertifikate über das **ACME-Protokoll** (Automated Certificate Management Environment) ausstellt.

**Zertifikat-Herausforderungen:**

| Challenge-Typ   | Beschreibung                                             | Wann geeignet                           |
| --------------- | -------------------------------------------------------- | --------------------------------------- |
| **HTTP-01**     | Datei unter `/.well-known/acme-challenge/` bereitstellen | Standard, einfach                       |
| **DNS-01**      | TXT-Record in DNS setzen                                 | Wildcard-Zertifikate, kein HTTP-Port 80 |
| **TLS-ALPN-01** | Spezielle TLS-Verbindung                                 | Wenn nur Port 443 verfügbar             |

**Automatische Erneuerung:** Let's Encrypt-Zertifikate sind 90 Tage gültig. Certbot, Traefik, Caddy und nginx mit Plugins erneuern automatisch.

### 9.4 HSTS (HTTP Strict Transport Security)

HSTS weist Browser an, eine Domain **ausschließlich über HTTPS** zu kontaktieren:

```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```

- `max-age`: Wie lange der Browser HSTS cachiert (1 Jahr empfohlen)
- `includeSubDomains`: Gilt für alle Subdomains
- `preload`: Aufnahme in Browser-HSTS-Preload-Liste (Domain immer HTTPS, auch ohne vorherigen Visit)

---

## 10. Security Testing

### 10.1 SAST — Static Application Security Testing

SAST analysiert den **Quellcode** auf Sicherheitslücken ohne die Anwendung auszuführen.

**Tools:**

| Tool                          | Sprachen             | Beschreibung                        |
| ----------------------------- | -------------------- | ----------------------------------- |
| **CodeQL**                    | Go, JS, Java, Python | GitHub-native, semantische Analyse  |
| **Semgrep**                   | Viele Sprachen       | Pattern-basiert, anpassbare Regeln  |
| **gosec**                     | Go                   | Go-spezifische Sicherheitsregeln    |
| **ESLint (security plugins)** | TypeScript           | Plugin-basiert, XSS, eval-Erkennung |

**Integration:** SAST sollte in der CI/CD-Pipeline als eigener Job laufen und bei neuen Findings den Build blockieren (oder PR-Review erzwingen).

### 10.2 DAST — Dynamic Application Security Testing

DAST testet die **laufende Anwendung** von außen — wie ein Angreifer.

**Tools:**

| Tool           | Typ         | Beschreibung                             |
| -------------- | ----------- | ---------------------------------------- |
| **OWASP ZAP**  | Open Source | Automatisierter Web-Scanner, Proxy-Modus |
| **Burp Suite** | Commercial  | Mächtiges Penetration-Testing-Tool       |
| **Nikto**      | Open Source | Schneller Web-Server-Scanner             |

**Einsatz:** In einer Test-/Staging-Umgebung nach jedem Deployment ausführen. DAST findet Lücken, die SAST übersieht (Konfigurationsfehler, Runtime-Verhalten).

### 10.3 Dependency Scanning

Bibliotheken und Dependencies enthalten bekannte Schwachstellen (CVEs). Dependency Scanning vergleicht verwendete Versionen mit CVE-Datenbanken.

**Tools:**

| Tool            | Ökosystem          | Integration                       |
| --------------- | ------------------ | --------------------------------- |
| **Dependabot**  | npm, Go, pip, etc. | GitHub-native, automatische PRs   |
| **Snyk**        | Viele              | CI/CD + IDE + Container           |
| **Trivy**       | Container, IaC     | Aqua Security, sehr schnell       |
| **govulncheck** | Go                 | Offizielles Go Vulnerability Tool |
| **npm audit**   | Node.js            | In npm eingebaut                  |

### 10.4 Penetration Testing

Penetration Testing ist der manuelle oder semi-automatische Versuch, eine Anwendung zu kompromittieren — wie ein echter Angreifer.

**Phasen:**

1. **Reconnaissance:** Informationen sammeln (DNS, Technologie-Stack, öffentliche Endpunkte)
2. **Scanning:** Automatisiertes Scanning (DAST, Port-Scanning)
3. **Exploitation:** Manuelle Ausnutzung gefundener Schwachstellen
4. **Post-Exploitation:** Was ist der potenzielle Schaden? Lateral Movement?
5. **Reporting:** Dokumentation mit Schweregrad und Empfehlungen

**Wann:** Vor Release größerer Features, nach Security-relevanten Änderungen, regelmäßig (jährlich, quartalsweise für kritische Systeme).

---

## 11. Entscheidungsmatrix: Auth-Pattern nach Anwendungsfall

| Anwendungsfall                     | Empfehlung                                        | Begründung                                                    |
| ---------------------------------- | ------------------------------------------------- | ------------------------------------------------------------- |
| **Klassische Web-App (SSR)**       | Session + httpOnly Cookie                         | Server-State einfach, gutes CSRF-Schutzpotenzial              |
| **SPA (React, Vue, Angular)**      | JWT (Access + Refresh Token) oder httpOnly Cookie | Stateless Backend; JWT im httpOnly Cookie empfohlen           |
| **Mobile App (iOS, Android)**      | OAuth2 PKCE + Token-Rotation                      | Keine Cookies, sichere Token-Speicherung im Keychain/Keystore |
| **API (M2M, keine Nutzer)**        | API Keys oder Client Credentials Flow             | Kein interaktiver Login nötig                                 |
| **Multi-Tenant SaaS**              | OIDC (externe IdP-Integration)                    | Delegierung an IdP, SSO für Enterprise-Kunden                 |
| **Kritische interne Apps (Admin)** | OIDC + MFA erzwingen                              | Höchste Sicherheitsanforderungen, SSO-Integration             |
| **Non-Profit / Vereinssoftware**   | JWT HS256, einfache RBAC                          | Einfachheit über Komplexität; keine externe IdP-Abhängigkeit  |

**Zusätzliche Kriterien:**

| Kriterium                   | Session                     | JWT                                      | OIDC                        |
| --------------------------- | --------------------------- | ---------------------------------------- | --------------------------- |
| **Skalierung**              | Session-Store nötig         | Stateless — einfach                      | Abhängig vom IdP            |
| **Revokierung**             | Sofort möglich              | Erst bei Token-Ablauf (→ kurze Laufzeit) | Token Introspection möglich |
| **Implementierungsaufwand** | Gering                      | Mittel                                   | Hoch (IdP-Integration)      |
| **Externe Dependencies**    | Keine (außer Session-Store) | JWT-Library                              | IdP (Auth0, Keycloak, etc.) |

---

## 12. Referenzen

### OWASP

- [OWASP Top 10 (2021)](https://owasp.org/Top10/2021/) — Die zehn kritischsten Web-Security-Risiken
- [OWASP Top 10 (2025)](https://owasp.org/Top10/2025/) — Aktuelle Version
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/) — Praxis-Checklisten für jeden Security-Aspekt
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) — Argon2id, bcrypt, Salting, Peppering
- [OWASP API Security Top 10](https://owasp.org/API-Security/) — API-spezifische Risiken
- [OWASP JWT Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html) — JWT-Sicherheitslücken und Mitigationen

### Authentifizierung & Autorisierung

- [RFC 7519 — JSON Web Token (JWT)](https://datatracker.ietf.org/doc/html/rfc7519) — JWT-Standard
- [RFC 7636 — PKCE](https://datatracker.ietf.org/doc/html/rfc7636) — Proof Key for Code Exchange
- [Google Zanzibar Paper (2019)](https://research.google/pubs/zanzibar-googles-consistent-global-authorization-system/) — ReBAC-Grundlage
- [OpenFGA Docs](https://openfga.dev/) — Open-Source Zanzibar-Implementation
- [Open Policy Agent](https://www.openpolicyagent.org/) — Policy-based Authorization
- [NIST SP 800-63B — Digital Identity Guidelines](https://pages.nist.gov/800-63-3/) — Passwort-Policies, MFA, Identity Proofing

### Standards & Browser-Security

- [MDN Web Security](https://developer.mozilla.org/en-US/docs/Web/Security) — CORS, CSP, SRI, HSTS aus Browser-Perspektive
- [W3C WebAuthn Specification](https://www.w3.org/TR/webauthn/) — Passkeys, FIDO2
- [Let's Encrypt Docs](https://letsencrypt.org/docs/) — ACME-Protokoll, automatische Zertifikatsverwaltung
- [12-Factor App — Config](https://12factor.net/config) — Environment Variables für Konfiguration und Secrets

### Bücher & Weiterführendes

- **Adam Shostack** (2014): _Threat Modeling: Designing for Security_ — STRIDE, umfassende Einführung
- **Michael Howard & David LeBlanc** (2002): _Writing Secure Code_ — Klassiker zu Secure Coding
- **OWASP Testing Guide** — Praktischer Leitfaden für Security Testing
