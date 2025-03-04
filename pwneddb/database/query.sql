-- name: GetPrefix :one
SELECT * FROM prefixes WHERE prefix = ? LIMIT 1;

-- name: GetEtags :many
SELECT id, prefix, etag FROM prefixes;

-- name: CreatePrefix :one
INSERT INTO prefixes (prefix, updated, etag, hashes) VALUES (?, ?, ?, ?) RETURNING *;

-- name: UpdatePrefix :exec
UPDATE prefixes set prefix = ?, updated = ?, etag = ?, hashes = ? WHERE id = ?;

-- name: DeletePrefix :exec
DELETE FROM prefixes WHERE id = ?;
