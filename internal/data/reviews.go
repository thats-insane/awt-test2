package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/thats-insane/awt-test1/internal/validator"
)

type Review struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	Author       string    `json:"author"`
	Rating       int64     `json:"rating"`
	HelpfulCount int32     `json:"helpful_count"`
	CreatedAt    time.Time `json:"-"`
}

type ReviewModel struct {
	DB *sql.DB
}

func (r ReviewModel) Insert(review *Review) error {
	query := `
	INSERT INTO reviews (product_id, author, rating, helpful_count)
	VALUES ($1, $2, $3, $4, COALESCE($5, 0))
	RETURNING id, created_at, version
	`
	args := []any{review.ProductID, review.Author, review.Rating, review.HelpfulCount}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&review.ID)
}

func (r ReviewModel) Get(id int64) (*Review, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
	SELECT id, product_id, author, rating, text, helpful_count, created_at, version
	FROM reviews
	WHERE id = $1
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := r.DB.QueryRowContext(ctx, query, id).Scan(&review.ID, &review.ProductID, &review.Author, &review.Rating, &review.HelpfulCount, &review.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &review, nil
}

func (r ReviewModel) GetAll(author string, rating string, helpfulCount string, filters Filters) ([]*Review, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), id, product_id, author, rating, helpful_count, created_at
	FROM reviews
	WHERE (to_tsvector('simple', author) @@
		plainto_tsquery('simple', $1) OR $1 = '') 
	AND (to_tsvector('simple', rating) @@
		plainto_tsquery('simple', $2) OR $2 = '') 
	AND (to_tsvector('simple', helpful_count) @@
		plainto_tsquery('simple', $3) OR $3 = '') 
	ORDER BY %s %s, id ASC 
	LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, query, author, rating, helpfulCount, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	reviews := []*Review{}

	for rows.Next() {
		var review Review
		err := rows.Scan(&totalRecords, &review.ID, &review.ProductID, &review.Author, &review.Rating, &review.HelpfulCount, &review.CreatedAt)
		if err != nil {
			return nil, Metadata{}, err
		}

		reviews = append(reviews, &review)
	}

	err = rows.Err()
	if err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return reviews, metadata, nil
}

func (r ReviewModel) Update(review *Review) error {
	query := `
	UPDATE reviews
	SET author = $1, rating = $2, helpful_count = $3
	WHERE id = $4
	RETURNING id
	`

	args := []any{review.Author, review.Rating, review.HelpfulCount, review.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.DB.QueryRowContext(ctx, query, args...).Scan(&review.ID)
}

func (r ReviewModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
	DELETE FROM reviews
	WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := r.DB.ExecContext(ctx, query, id)
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

func (r *ReviewModel) Exists(id int64) (bool, error) {
	query := `
	SELECT EXISTS
	(SELECT 1 FROM reviews WHERE id = $1)
	`
	var exists bool

	err := r.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Author != "", "author", "must be provided")
	v.Check(len(review.Author) <= 25, "author", "must not be more than 25 bytes long")
	v.Check(review.ProductID > 0, "product_id", "must be a positive integer")
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5")
}
