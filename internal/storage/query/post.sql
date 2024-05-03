-- name: InsertPost :one
INSERT INTO post (title, description, num_images, profile_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: GetPost :one
SELECT *
FROM post
WHERE id = $1;
