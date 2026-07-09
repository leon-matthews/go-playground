-- name: GetHashCount :one
SELECT count FROM hashes WHERE hash = ?;

-- name: UpsertPassword :execrows
INSERT INTO passwords (password, count) VALUES (?, ?)
ON CONFLICT(password) DO UPDATE SET count = excluded.count
WHERE count != excluded.count;

-- name: InsertPassword :execrows
INSERT INTO passwords (password, count) VALUES (?, ?)
ON CONFLICT(password) DO NOTHING;

-- name: TopPasswords :many
SELECT password, count FROM passwords ORDER BY count DESC LIMIT ?;
