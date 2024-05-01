-- name: InsertIngredient :one
INSERT INTO ingredient (name)
VALUES ($1)
RETURNING id;

-- name: GetIngredientId :one
SELECT id
FROM ingredient
WHERE name = $1;

-- name: GetSimilarIngredient :many
SELECT name
FROM ingredient
WHERE name LIKE '%' || sqlc.arg(name) || '%'
ORDER BY usage DESC
LIMIT 10;	
