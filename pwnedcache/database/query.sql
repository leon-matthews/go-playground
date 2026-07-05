-- name: GetPrefix :one
SELECT * FROM prefixes WHERE prefix = ? LIMIT 1;

-- name: GetEtags :many
SELECT id, prefix, etag FROM prefixes;

-- name: UpsertPrefix :exec
INSERT INTO prefixes (prefix, updated, etag) VALUES (?, ?, ?)
ON CONFLICT(prefix) DO UPDATE SET
  updated = excluded.updated,
  etag = excluded.etag;

-- name: DeletePrefix :exec
DELETE FROM prefixes WHERE id = ?;

-- name: InsertHash :exec
INSERT INTO hashes (hash, count) VALUES (?, ?);

-- name: GetHashCount :one
SELECT count FROM hashes WHERE hash = ?;

-- name: DeleteHashRange :exec
DELETE FROM hashes WHERE hash BETWEEN sqlc.arg(lower) AND sqlc.arg(upper);
