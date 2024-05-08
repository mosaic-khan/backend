-- name: LikePost :exec
INSERT INTO profile_like_post (profile_id, post_id)
VALUES ($1, $2);

-- name: DislikePost :exec
DELETE
FROM profile_like_post
WHERE post_id = $1 AND profile_id = $2;

-- name: ProfileLikePost :one
SELECT COUNT(*)
FROM profile_like_post
WHERE profile_id = $1 AND post_id = $2;