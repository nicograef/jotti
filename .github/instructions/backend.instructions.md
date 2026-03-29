---
description: "Use when working on Go backend code, API handlers, middleware, repositories, domain models, or application services."
applyTo: "backend/**"
---

> **Referenz:** Für Ubiquitous Language und Namenskonventionen pro Schicht → `docs/language.md`. Für Architektur, Invarianten und Schichtenarchitektur → `docs/handbuch.md` §6.

Repo-weite Regeln und Guardrails stehen kanonisch in `AGENTS.md`. Diese Datei ergänzt nur backend-spezifische Konventionen für `backend/**`.

# Backend-Konventionen

## Befehle

- **Build:** `make build-backend`
- **Unit-Tests:** `make test`
- **Lint:** `make lint-backend`
- **Format:** `make fmt-backend`
- **sqlc generieren:** `make sqlc` (nach Query-Änderungen)

## Verzeichnisstruktur

```
backend/
  main.go                       # Einstiegspunkt
  sqlc.yaml                     # sqlc-Konfiguration
	sqlc/queries/                 # SQL-Queries für sqlc
	sqlc/dbgen/                   # Generierter Code
  api/service.go                # Service-Routen (Kasse — Tisch-Operationen)
  api/serviceleitung.go         # serviceleitung-Routen (Stornierung)
  api/admin.go                  # Admin-Routen (Verwaltung, Kassensitzung)
  api/auth.go                   # Auth-Routen (Login, Passwort setzen)
  api/<domain>/http/            # HTTP-Handler
  api/<domain>/application/     # Application-Services
  api/middleware/               # JWT-Auth, Rate-Limiting, Logging
  api/helper/                   # HTTP-Hilfsfunktionen (JSON-Parsing, Response)
  domain/<domain>/              # Domain-Modelle und Business-Logik
  repository/<domain>_repo/     # Datenbank-Zugriff (sqlc-basiert)
  config/                       # Konfiguration aus Umgebungsvariablen
  app/                          # App-Struct (Dependency Wiring)
  db/                           # Datenbank-Verbindung und Fehler-Mapping
```

## Fehlerformat

Alle Fehler-Responses: `{"code": "<string>", "details": "<optional>"}` (siehe `api/helper/http.go`).

## Auth

- JWT HS256, 12h Gültigkeit, Claims: `sub` (userID), `role` (admin|serviceleitung|service)
- Middleware extrahiert `userID` und `role` aus JWT in Request-Context
- Passwörter: Argon2id-Hashing (`domain/user/password.go`)

## Schichtentrennung: Domain vs. HTTP

> Grundregel: siehe AGENTS.md Regel #10. Hier das konkrete Pattern:

Beispiel fuer Response-DTO und Mapper in der HTTP-Schicht:

```go
type varianteDTO struct {
	ID         int       `json:"id"`
	Name       string    `json:"name"`
	PreisCents int       `json:"preisCents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func toVarianteDTO(v product.Variante) varianteDTO {
	return varianteDTO{
		ID:         v.ID,
		Name:       v.Name,
		PreisCents: v.PreisCents,
		Status:     string(v.Status),
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	}
}

type produktDTO struct {
	ID        int           `json:"id"`
	Name      string        `json:"name"`
	Kategorie string        `json:"kategorie"`
	Status    string        `json:"status"`
	Varianten []varianteDTO `json:"varianten"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

func toProduktDTO(p product.Produkt) produktDTO {
	varianten := make([]varianteDTO, 0, len(p.Varianten))
	for _, variante := range p.Varianten {
		varianten = append(varianten, toVarianteDTO(variante))
	}

	return produktDTO{
		ID:        p.ID,
		Name:      p.Name,
		Kategorie: string(p.Kategorie),
		Status:    string(p.Status),
		Varianten: varianten,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
```

## Tests

Unit-Tests mit `//go:build unit` Tag. Ausführen: `make test`

## Weiterführende Dokumentation

- **Schichtenarchitektur, API-Design, Validierung, OCC:** [docs/handbuch.md](../../docs/handbuch.md) Kap. 6
- **Namenskonventionen (Deutsch/Englisch pro Schicht, Ist vs. Soll):** [docs/language.md](../../docs/language.md)

## Code-Beispiele

### HTTP-Handler

Pattern: Request-Struct definieren → Body lesen → Service aufrufen → Fehler mappen → Response senden.

Hinweis: Query-Handler senden keine Domain-Modelle direkt, sondern mappen immer auf Response-DTOs.

```go
type createProdukt struct {
	Name      string           `json:"name"`
	Kategorie product.Kategorie `json:"kategorie"`
}

type createProduktResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateProduktHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createProdukt{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		id, err := h.Command.CreateProdukt(r.Context(), body.Name, body.Kategorie)
		if err != nil {
			switch {
			case errors.Is(err, application.ErrProduktAlreadyExists):
				helper.SendClientError(w, "produkt_already_exists", nil)
			case errors.Is(err, application.ErrInvalidProduktData):
				helper.SendClientError(w, "invalid_produkt_data", nil)
			default:
				helper.SendServerError(w)
			}
			return
		}

		// Query-Handler folgen demselben Prinzip: Domain -> Response-DTO vor SendResponse.
		helper.SendResponse(w, createProduktResponse{ID: id})
	}
}
```

### Application-Service

Pattern: Logging → Domain-Modell aufbauen/validieren → Repository aufrufen → Fehler mappen.

```go
func (c Command) CreateProdukt(ctx context.Context, name string, kategorie product.Kategorie) (int, error) {
	log := zerolog.Ctx(ctx)

	produkt, err := product.NewProdukt(name, kategorie)
	if err != nil {
		log.Warn().Err(err).Str("name", name).Msg("Invalid produkt data")
		return 0, ErrInvalidProduktData
	}

	produktID, err := c.ProduktRepo.CreateProdukt(ctx, produkt)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", produkt.Name).Msg("Produkt name already exists")
			return 0, ErrProduktAlreadyExists
		} else {
			log.Error().Str("name", produkt.Name).Msg("Failed to create produkt")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("produkt_id", produktID).Msg("Produkt created")
	return produktID, nil
}
```

### zog-Validierungsschema

```go
var NameSchema = z.String().Trim().Min(3, z.Message("Name zu kurz")).Max(100, z.Message("Name zu lang"))

var KategorieSchema = z.StringLike[Kategorie]().OneOf(
	[]Kategorie{EssenKategorie, GetraenkKategorie, SonstigesKategorie},
	z.Message("Ungültige Kategorie"),
)

var ProduktSchema = z.Struct(z.Shape{
	"ID":        IDSchema.Required(),
	"Name":      NameSchema.Required(),
	"Kategorie": KategorieSchema.Required(),
	"Varianten": z.Slice(VarianteSchema).Required(),
	"CreatedAt": z.Time().Required(),
})

func NewProdukt(name string, kategorie Kategorie) (Produkt, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Produkt{}, fmt.Errorf("invalid name")
	}
	if issue := KategorieSchema.Validate(&kategorie); issue != nil {
		return Produkt{}, fmt.Errorf("invalid category")
	}
	return Produkt{Name: name, Kategorie: kategorie}, nil
}
```
