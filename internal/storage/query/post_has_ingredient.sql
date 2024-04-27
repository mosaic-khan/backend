-- name: InsertPostHasIngredient :exec
INSERT INTO post_has_ingredient (post_id, ingredient_id, amount)
VALUES ($1, $2, $3);

-- name: GetPostIngredient :many
SELECT i.name, p.amount
FROM post_has_ingredient p JOIN ingredient i ON p.ingredient_id = i.id
WHERE p.post_id = $1;