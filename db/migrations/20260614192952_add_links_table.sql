-- +goose Up
CREATE TABLE IF NOT EXISTS links(
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  original_url TEXT NOT NULL CONSTRAINT links_original_url_unique UNIQUE,
  short_name VARCHAR(50) NOT NULL CONSTRAINT links_short_name_unique UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE links;
