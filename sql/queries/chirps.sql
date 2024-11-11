-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(), NOW(), NOW(), $1, $2
)
RETURNING *;

-- name: GetChirps :many
SELECT * FROM chirps
ORDER BY
    CASE WHEN @order_by::TEXT = 'ASC' THEN chirps.created_at END ASC,
    CASE WHEN @order_by::TEXT = 'DESC' THEN chirps.created_at END DESC;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE id = $1;

-- name: UpgradeUser :one
UPDATE users
SET is_chirpy_red = TRUE
WHERE id = $1
RETURNING *;

-- name: GetChirpsByAuthor :many
SELECT * FROM chirps
WHERE user_id = $1
ORDER BY
    CASE WHEN @order_by::TEXT = 'ASC' THEN chirps.created_at END ASC,
    CASE WHEN @order_by::TEXT = 'DESC' THEN chirps.created_at END DESC;
