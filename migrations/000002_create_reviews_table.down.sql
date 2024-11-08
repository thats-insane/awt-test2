CREATE TABLE IF NOT EXISTS (
    id bigserial PRIMARY KEY,
	product_id bigserial REFERENCES products,
	author text NOT NULL,
	rating integer NOT NULL,
	helpful_count integer NOT NULL,
	created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
)