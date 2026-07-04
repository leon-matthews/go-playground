-- name: GetHashCount :one
SELECT count FROM hashes WHERE hash = ?;

-- name: UpsertPassword :exec
INSERT INTO passwords (password, count) VALUES (?, ?)
ON CONFLICT(password) DO UPDATE SET count = excluded.count;

-- name: TopPasswords :many
SELECT password, count FROM passwords ORDER BY count DESC LIMIT ?;
