package product

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"
)

// Persistence implements product persistence layer using a SQL database.
type Persistence struct {
	DB *sql.DB
}

type dbproduct struct {
	ID          int          `db:"id"`
	Name        string       `db:"name"`
	Description string       `db:"description"`
	NetPrice    float64      `db:"net_price"`
	Status      string       `db:"status"`
	Category    string       `db:"category"`
	CreatedAt   sql.NullTime `db:"created_at"`
}

// GetProduct retrieves a product from the database by its ID.
func (p *Persistence) GetProduct(ctx context.Context, id int) (*Product, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	row := p.DB.QueryRowContext(ctx, "SELECT id, name, description, net_price, status, category, created_at FROM products WHERE id = $1 AND status != 'deleted'", id)

	var dbProduct dbproduct
	if err := row.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPrice, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			log.Warn().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product not found")
			return nil, ErrProductNotFound
		}
		log.Error().Str("correlation_id", correlationID).Err(err).Int("product_id", id).Msg("Failed to scan product")
		return nil, err
	}

	log.Debug().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product retrieved")
	return &Product{
		ID:          dbProduct.ID,
		Name:        dbProduct.Name,
		Description: dbProduct.Description,
		NetPrice:    dbProduct.NetPrice,
		Status:      Status(dbProduct.Status),
		Category:    Category(dbProduct.Category),
		CreatedAt:   dbProduct.CreatedAt.Time,
	}, nil
}

// GetAllProducts retrieves all products from the database.
func (p *Persistence) GetAllProducts(ctx context.Context) ([]*Product, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, description, net_price, status, category, created_at FROM products WHERE status != 'deleted' ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var products []*Product
	for rows.Next() {
		var dbProduct dbproduct
		if err := rows.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPrice, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt); err != nil {
			return nil, err
		}

		products = append(products, &Product{
			ID:          dbProduct.ID,
			Name:        dbProduct.Name,
			Description: dbProduct.Description,
			NetPrice:    dbProduct.NetPrice,
			Status:      Status(dbProduct.Status),
			Category:    Category(dbProduct.Category),
			CreatedAt:   dbProduct.CreatedAt.Time,
		})
	}

	log.Debug().Str("correlation_id", correlationID).Int("count", len(products)).Msg("Retrieved all products")
	return products, nil
}

// GetActiveProducts retrieves active products from the database.
func (p *Persistence) GetActiveProducts(ctx context.Context) ([]*ProductPublic, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	rows, err := p.DB.QueryContext(ctx, "SELECT id, name, description, net_price, category FROM products WHERE status = 'active' ORDER BY category, name ASC")
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			err = closeErr
		}
	}()

	var products []*ProductPublic
	for rows.Next() {
		var dbProduct dbproduct
		if err := rows.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPrice, &dbProduct.Category); err != nil {
			return nil, err
		}

		products = append(products, &ProductPublic{
			ID:          dbProduct.ID,
			Name:        dbProduct.Name,
			Description: dbProduct.Description,
			NetPrice:    dbProduct.NetPrice,
			Category:    Category(dbProduct.Category),
		})
	}

	log.Debug().Str("correlation_id", correlationID).Int("count", len(products)).Msg("Retrieved active products")
	return products, nil
}

// CreateProduct inserts a new product into the database.
func (p *Persistence) CreateProduct(ctx context.Context, name, description string, netPrice float64, category Category) (int, error) {
	correlationID, _ := ctx.Value("correlation_id").(string)
	var id int
	err := p.DB.QueryRowContext(ctx,
		"INSERT INTO products (name, description, net_price, category) VALUES ($1, $2, $3, $4) RETURNING id",
		name, description, netPrice, string(category),
	).Scan(&id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Str("name", name).Msg("Failed to create product")
		return 0, err
	}
	log.Info().Str("correlation_id", correlationID).Int("product_id", id).Str("name", name).Msg("Product created")
	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (p *Persistence) UpdateProduct(ctx context.Context, id int, name, description string, netPrice float64, category Category) error {
	correlationID, _ := ctx.Value("correlation_id").(string)
	result, err := p.DB.ExecContext(ctx,
		"UPDATE products SET name = $1, description = $2, net_price = $3, category = $4 WHERE id = $5 AND status != 'deleted'",
		name, description, netPrice, string(category), id,
	)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Int("product_id", id).Msg("Failed to update product")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Warn().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product not found for update")
		return ErrProductNotFound
	}

	log.Info().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product updated")
	return nil
}

// ActivateProduct sets the status of a product to active.
func (p *Persistence) ActivateProduct(ctx context.Context, id int) error {
	correlationID, _ := ctx.Value("correlation_id").(string)
	result, err := p.DB.ExecContext(ctx, "UPDATE products SET status = 'active' WHERE id = $1 AND status != 'deleted'", id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Int("product_id", id).Msg("Failed to activate product")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Warn().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product not found for activation")
		return ErrProductNotFound
	}

	log.Info().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product activated")
	return nil
}

// DeactivateProduct sets the status of a product to inactive.
func (p *Persistence) DeactivateProduct(ctx context.Context, id int) error {
	correlationID, _ := ctx.Value("correlation_id").(string)
	result, err := p.DB.ExecContext(ctx, "UPDATE products SET status = 'inactive' WHERE id = $1 AND status != 'deleted'", id)
	if err != nil {
		log.Error().Str("correlation_id", correlationID).Err(err).Int("product_id", id).Msg("Failed to deactivate product")
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		log.Warn().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product not found for deactivation")
		return ErrProductNotFound
	}

	log.Info().Str("correlation_id", correlationID).Int("product_id", id).Msg("Product deactivated")
	return nil
}
