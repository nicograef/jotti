package product

import (
	"database/sql"
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
func (p *Persistence) GetProduct(id int) (*Product, error) {
	row := p.DB.QueryRow("SELECT id, name, description, net_price, status, category, created_at FROM products WHERE id = $1 AND status != 'deleted'", id)

	var dbProduct dbproduct
	if err := row.Scan(&dbProduct.ID, &dbProduct.Name, &dbProduct.Description, &dbProduct.NetPrice, &dbProduct.Status, &dbProduct.Category, &dbProduct.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

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
func (p *Persistence) GetAllProducts() ([]*Product, error) {
	rows, err := p.DB.Query("SELECT id, name, description, net_price, status, category, created_at FROM products WHERE status != 'deleted' ORDER BY id ASC")
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

	return products, nil
}

// CreateProduct inserts a new product into the database.
func (p *Persistence) CreateProduct(name, description string, netPrice float64, category Category) (int, error) {
	var id int
	err := p.DB.QueryRow(
		"INSERT INTO products (name, description, net_price, category) VALUES ($1, $2, $3, $4) RETURNING id",
		name, description, netPrice, string(category),
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateProduct updates an existing product in the database.
func (p *Persistence) UpdateProduct(id int, name, description string, netPrice float64, category Category) error {
	result, err := p.DB.Exec(
		"UPDATE products SET name = $1, description = $2, net_price = $3, category = $4 WHERE id = $5 AND status != 'deleted'",
		name, description, netPrice, string(category), id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

// ActivateProduct sets the status of a product to active.
func (p *Persistence) ActivateProduct(id int) error {
	result, err := p.DB.Exec("UPDATE products SET status = 'active' WHERE id = $1 AND status != 'deleted'", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}

// DeactivateProduct sets the status of a product to inactive.
func (p *Persistence) DeactivateProduct(id int) error {
	result, err := p.DB.Exec("UPDATE products SET status = 'inactive' WHERE id = $1 AND status != 'deleted'", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrProductNotFound
	}

	return nil
}
