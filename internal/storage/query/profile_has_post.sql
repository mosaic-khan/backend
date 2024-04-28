-- name: InsertProfilerHasPost :exec
INSERT INTO profile_has_post (profile_id, post_id)
VALUES ($1, $2);

-- name: GetProfilePostId :many
SELECT post_id
FROM profile_has_post
WHERE profile_id = $1;

-- name: GetAccountPostId :many
SELECT post_id
FROM profile p JOIN profile_has_post pp ON p.id = pp.profile_id
WHERE p.user_id = $1;
