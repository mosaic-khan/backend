-- name: InsertPost :one
INSERT INTO post (title, description, num_images)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetPost :one
SELECT *
FROM post
WHERE id = $1;
