package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/thats-insane/awt-test1/internal/validator"
)

var ErrRecordNotFound = errors.New("record not found")

type Product struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Price         float64   `json:"price"`
	AverageRating float64   `json:"average_rating"`
	ImageURL      string    `json:"image_url"`
	CreatedAt     time.Time `json:"created_at"`
}

type ProductModel struct {
	DB *sql.DB
}

func (p ProductModel) Insert(product *Product) error {
	query := `
	INSERT INTO products (name, description, category, price, average_rating, image_url) 
	VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING id, created_at
	`

	args := []any{product.Name, product.Description, product.Category, product.Price, 0, product.ImageURL}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID, &product.CreatedAt)
}

func (p ProductModel) Get(id int64) (*Product, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, name, description, category, price, average_rating, image_url, created_at
	FROM products
	WHERE id = $1;
	`

	var product Product

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, id).Scan(&product.ID, &product.Name, &product.Description, &product.Category, &product.Price, &product.AverageRating, &product.ImageURL, &product.CreatedAt)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &product, nil
}

func (p ProductModel) GetAll(name string, description string, category string, price string, avgRating string, imageURL string, filters Filters) ([]*Product, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), id, name, description, category, price, average_rating, image_url, created_at
	FROM products
	WHERE (to_tsvector('simple', name) @@
		plainto_tsquery('simple', $1) OR $1 = '')
    AND (to_tsvector('simple', description) @@
		plainto_tsquery('simple', $2) OR $2 = '')
	AND (to_tsvector('simple', category) @@
		plainto_tsquery('simple', $3) OR $3 = '')
	AND (to_tsvector('simple', price) @@
		plainto_tsquery('simple', $4) OR $4 = '')
	AND (to_tsvector('simple', average_rating) @@
		plainto_tsquery('simple', $5) OR $5 = '')
	AND (to_tsvector('simple', image_url) @@
		plainto_tsquery('simple', $6) OR $6 = '')
	ORDER BY %s %s, id ASC 
        LIMIT $7 OFFSET $8
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, name, description, category, price, avgRating, imageURL, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()
	totalRecords := 0
	products := []*Product{}

	for rows.Next() {
		var product Product
		err := rows.Scan(&totalRecords, &product.ID, &product.Name, &product.Description, &product.Category, &product.Price, &product.AverageRating, &product.ImageURL, &product.CreatedAt)
		if err != nil {
			return nil, Metadata{}, err
		}

		products = append(products, &product)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return products, metadata, nil
}

func (p ProductModel) Update(product *Product) error {

	query := `
	UPDATE products 
	SET name = $1, description = $2, category = $3, price = $4, image_url = $5, average_rating = $6 
	WHERE id = $7
	RETURNING id
	`

	args := []any{product.Name, product.Description, product.Category, product.Price, product.ImageURL, product.AverageRating, product.ID}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := p.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID)
	if err != nil {
		return err
	}

	return p.DB.QueryRowContext(ctx, query, args...).Scan(&product.ID)
}

func (p ProductModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
	DELETE FROM products
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (p ProductModel) Exists(productID int64) (bool, error) {
	query := `
	SELECT EXISTS
	(SELECT 1 FROM products WHERE id = $1)
	`
	var exists bool

	err := p.DB.QueryRow(query, productID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func ValidateProduct(v *validator.Validator, product *Product, handler int) {
	switch handler {
	case 1:
		v.Check(product.Name != "", "name", "must be provided")
		v.Check(product.Description != "", "description", "must be provided")
		v.Check(product.Category != "", "category", "must be provided")
		v.Check(product.AverageRating != 0, "average_rating", "must be provided")
		v.Check(product.ImageURL != "", "image_url", "must be provided")

		v.Check(len(product.Name) <= 100, "name", "must not be more than 100 byte long")
		v.Check(len(product.Description) <= 100, "description", "must not be more than 100 byte long")
		v.Check(len(product.Category) <= 100, "category", "must not be more than 100 byte long")
	default:
		log.Printf("Unable to locate handler ID: %d", handler)
		v.AddError("default", "Handler ID not provided")
	}
}
