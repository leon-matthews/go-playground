-- name: GetPrefix :one
SELECT * FROM prefixes WHERE prefix = ? LIMIT 1;

-- name: GetEtags :many
SELECT id, prefix, etag FROM prefixes;

-- name: UpsertPrefix :exec
INSERT INTO prefixes (prefix, updated, etag, hashes) VALUES (?, ?, ?, ?)
ON CONFLICT(prefix) DO UPDATE SET
  updated = excluded.updated,
  etag = excluded.etag,
  hashes = excluded.hashes;

-- name: DeletePrefix :exec
DELETE FROM prefixes WHERE id = ?;
