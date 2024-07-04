-- name: AddPin :exec
INSERT INTO profile_pin_post(profile_id, post_id)
VALUES ($1, $2);

-- name: RemovePin :exec
DELETE
FROM profile_pin_post
WHERE profile_id = $1 AND post_id = $2;

-- name: GetPinsCount :one
SELECT count(*)
FROM profile_pin_post
WHERE profile_id = $1;

-- name: GetPins :many
SELECT post.id, post.title, post_image.image_url
FROM profile_pin_post
    JOIN post ON profile_pin_post.post_id = post.id
    LEFT JOIN post_image ON profile_pin_post.post_id = post_image.post_id
WHERE profile_pin_post.profile_id = $1 AND post_image.is_primary = TRUE;
