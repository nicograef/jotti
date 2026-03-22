---
description: "Use when working on Go backend code, API handlers, middleware, repositories, domain models, or application services."
applyTo: "backend/**"
---

> **Referenz:** Für Ubiquitous Language und Namenskonventionen pro Schicht → `docs/language.md`. Für Architektur, Invarianten und Schichtenarchitektur → `docs/handbuch.md` §6.

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
  sqlc/dbgen/                   # Generierter Code (NICHT EDITIEREN)
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

- Domain-Structs in `domain/` tragen keine `json`-Tags. Die Domain kennt kein HTTP.
- Die HTTP-Schicht (`api/<domain>/http/`) definiert eigene Request- und Response-DTOs mit `json`-Tags.
- Query-Handler mappen Domain-Modelle in Response-DTOs, bevor sie `helper.SendResponse` aufrufen.
- Command-Handler nutzen Request-DTOs fuer Request-Bodies und geben Response-DTOs zurueck.
- Ausnahme: Event-Data-Structs fuer Event-Store-Persistenz duerfen `json`-Tags tragen.

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
type createProduct struct {
	Name     string           `json:"name"`
	Category product.Category `json:"category"`
}

type createProductResponse struct {
	ID int `json:"id"`
}

func (h *CommandHandler) CreateProductHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := createProduct{}
		if !helper.ReadBody(w, r, &body) {
			return
		}

		id, err := h.Command.CreateProduct(r.Context(), body.Name, body.Category)
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
		helper.SendResponse(w, createProductResponse{ID: id})
	}
}
```

### Application-Service

Pattern: Logging → Domain-Modell aufbauen/validieren → Repository aufrufen → Fehler mappen.

```go
func (c Command) CreateProduct(ctx context.Context, name string, category product.Category) (int, error) {
	log := zerolog.Ctx(ctx)

	product, err := product.NewProduct(name, category)
	if err != nil {
		log.Warn().Err(err).Str("product_name", name).Msg("Invalid product data")
		return 0, ErrInvalidProduktData
	}

	productID, err := c.ProductRepo.CreateProduct(ctx, product)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			log.Warn().Err(err).Str("name", product.Name).Msg("Product name already exists")
			return 0, ErrProduktAlreadyExists
		} else {
			log.Error().Str("name", product.Name).Msg("Failed to create product")
			return 0, ErrDatabase
		}
	}

	log.Info().Int("product_id", productID).Msg("Product created")
	return productID, nil
}
```

### zog-Validierungsschema

```go
var NameSchema = z.String().Trim().Min(3, z.Message("Name too short")).Max(100, z.Message("Name too long"))

var CategorySchema = z.StringLike[Category]().OneOf(
	[]Category{FoodCategory, BeverageCategory, OtherCategory},
	z.Message("Invalid category"),
)

var ProductSchema = z.Struct(z.Shape{
	"ID":        IDSchema.Required(),
	"Name":      NameSchema.Required(),
	"Category":  CategorySchema.Required(),
	"Variants":  z.Slice(VariantSchema).Required(),
	"CreatedAt": z.Time().Required(),
})

func NewProduct(name string, category Category) (Product, error) {
	if issue := NameSchema.Validate(&name); issue != nil {
		return Product{}, fmt.Errorf("invalid name")
	}
	if issue := CategorySchema.Validate(&category); issue != nil {
		return Product{}, fmt.Errorf("invalid category")
	}
	return Product{Name: name, Category: category}, nil
}
```
