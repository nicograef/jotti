package product_repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nicograef/jotti/backend/db"
	"github.com/nicograef/jotti/backend/domain/product"
)

// scanProductWithVariants scans a row containing product fields and a JSON variants array,
// then converts them to a domain Product with nested Variants.
func scanProductWithVariants(scan func(...any) error) (product.Product, error) {
	var p dbproduct
	var variantsJSON []byte

	if err := scan(&p.ID, &p.Name, &p.Category, &p.CreatedAt, &variantsJSON); err != nil {
		return product.Product{}, db.Error(err)
	}

	prod := p.toDomain()

	var variants []dbvariant
	if err := json.Unmarshal(variantsJSON, &variants); err != nil {
		return product.Product{}, fmt.Errorf("failed to unmarshal variants: %w", err)
	}

	prod.Variants = make([]product.Variant, 0, len(variants))
	for _, v := range variants {
		prod.Variants = append(prod.Variants, v.toDomain())
	}

	return prod, nil
}

func (r Repository) GetProduct(ctx context.Context, id int) (product.Product, error) {
	query := `
		SELECT 
			p.id,
			p.name,
			p.category,
			p.created_at,
			COALESCE(
				(SELECT json_agg(
					json_build_object(
						'id', pv.id,
						'name', pv.name,
						'price_cents', pv.price_cents,
						'status', pv.status,
						'created_at', pv.created_at
					)
				)
				FROM product_variants pv
				WHERE pv.product_id = p.id AND pv.status != 'deleted'),
				'[]'
			) AS variants
		FROM products p
		WHERE p.id = $1
	`

	row := r.DB.QueryRowContext(ctx, query, id)
	return scanProductWithVariants(row.Scan)
}

func (r Repository) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	query := `
		WITH variant_json AS (
			SELECT 
				product_id,
				json_agg(
					json_build_object(
						'id', id,
						'name', name,
						'price_cents', price_cents,
						'status', status,
						'created_at', created_at
					)
				) AS variants
			FROM product_variants
			WHERE status != 'deleted'
			GROUP BY product_id
		)
		SELECT 
			p.id,
			p.name,
			p.category,
			p.created_at,
			COALESCE(vj.variants, '[]') AS variants
		FROM products p
		LEFT JOIN variant_json vj ON vj.product_id = p.id
		ORDER BY p.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "products with variants")

	products := []product.Product{}
	for rows.Next() {
		prod, err := scanProductWithVariants(rows.Scan)
		if err != nil {
			return nil, err
		}
		products = append(products, prod)
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return products, nil
}

func (r Repository) GetActiveProducts(ctx context.Context) ([]product.Product, error) {
	query := `
		WITH variant_json AS (
			SELECT 
				product_id,
				json_agg(
					json_build_object(
						'id', id,
						'name', name,
						'price_cents', price_cents,
						'status', status,
						'created_at', created_at
					)
				) AS variants
			FROM product_variants
			WHERE status = 'active'
			GROUP BY product_id
		)
		SELECT 
			p.id,
			p.name,
			p.category,
			p.created_at,
			vj.variants
		FROM products p
		INNER JOIN variant_json vj ON vj.product_id = p.id
		ORDER BY p.id ASC
	`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, db.Error(err)
	}
	defer db.Close(rows, "active products with variants")

	products := []product.Product{}
	for rows.Next() {
		prod, err := scanProductWithVariants(rows.Scan)
		if err != nil {
			return nil, err
		}
		products = append(products, prod)
	}

	if err := rows.Err(); err != nil {
		return nil, db.Error(err)
	}

	return products, nil
}

func (r Repository) CreateProduct(ctx context.Context, p product.Product) (int, error) {
	var id int
	err := r.DB.QueryRowContext(ctx,
		"INSERT INTO products (name, category, created_at) VALUES ($1, $2, $3) RETURNING id",
		p.Name, string(p.Category), p.CreatedAt,
	).Scan(&id)
	if err != nil {
		return 0, db.Error(err)
	}

	return id, nil
}

func (r Repository) UpdateProduct(ctx context.Context, p product.Product) error {
	result, err := r.DB.ExecContext(ctx,
		"UPDATE products SET name = $1, category = $2 WHERE id = $3",
		p.Name, string(p.Category), p.ID,
	)
	if err != nil {
		return db.Error(err)
	}

	return db.ResultError(result)
}
