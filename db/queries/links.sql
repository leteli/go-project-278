-- name: GetLinks :many
SELECT id, original_url, short_name, created_at
FROM links
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateLink :one
INSERT INTO links
(original_url, short_name)
VALUES ($1, $2)
RETURNING id, original_url, short_name, created_at;

-- name: GetLink :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE id = $1;

-- name: UpdateLink :one
UPDATE links
SET
  original_url = COALESCE(sqlc.narg(original_url), original_url),
  short_name = COALESCE(sqlc.narg(short_name), short_name)
WHERE id = sqlc.arg(id)
RETURNING id, original_url, short_name, created_at;

-- name: DeleteLink :execrows
DELETE FROM links
WHERE id = $1;