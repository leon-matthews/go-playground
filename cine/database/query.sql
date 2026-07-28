-- Starter queries. The bulk-import hot path uses hand-written batched INSERTs
-- (see the importer); these exercise codegen across the main table shapes.

-- name: GetTitle :one
SELECT * FROM titles WHERE id = ? LIMIT 1;

-- name: InsertTitle :exec
INSERT INTO titles (
  id, title_type, primary_title, original_title, is_adult,
  start_year, end_year, runtime_minutes, genres, average_rating, num_votes
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpsertGenre :one
INSERT INTO titles_genre (id, name) VALUES (?, ?)
ON CONFLICT(id) DO UPDATE SET name = excluded.name
RETURNING id;

-- name: ListPrincipalsForTitle :many
SELECT * FROM principals WHERE title_id = ? ORDER BY ordering;
