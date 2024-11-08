CREATE TABLE IF NOT EXISTS products (
    id bigserial PRIMARY KEY,
	name text NOT NULL,
	description text NOT NULL,
	category text NOT NULL,
	price bigint NOT NULL,
	average_rating bigint NOT NULL,
	image_url text NOT NULL,
	created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
);