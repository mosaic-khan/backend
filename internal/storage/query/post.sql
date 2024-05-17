-- name: InsertPost :one
INSERT INTO post (title, description, category_id, num_images, profile_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetPost :one
SELECT post.id, post.title, post.description, category.name as category, account.username, profile.profile_pic_address, post.num_images, post.num_likes
FROM post
         JOIN profile on profile.id = post.profile_id
         JOIN account on account.id = profile.user_id
         JOIN category on post.category_id = category.id
WHERE post.id = $1;

-- name: PostImageCount :one
SELECT count(*)
FROM post_image
WHERE post_id = $1;

-- name: AddImage :exec
WITH post_image_cnt AS (
    SELECT COUNT(*) as cnt
    FROM post_image
    WHERE post_id = $1
)
INSERT INTO post_image (post_id, image_url, is_primary)
SELECT $1, $2, (cnt < 1)::boolean
FROM post_image_cnt;


-- name: GetPostImages :many
SELECT post_image.image_url
FROM post_image
WHERE post_id = $1
ORDER BY id;

-- name: GetPostsPreview :many
SELECT post.id, post.title, post.description, post_image.image_url
FROM post
    LEFT JOIN post_image on post.id = post_image.post_id
WHERE post.profile_id = $1 and post_image.is_primary = true
ORDER BY post.id
LIMIT 20;


-- name: GetPostOwnerProfile :one
SELECT post.profile_id AS profile_id
FROM post
WHERE id = $1;

-- name: GetPostsWithCategory :many
SELECT post.id, title, description, post_image.image_url as post_image, username, profile_pic_address
FROM post
	INNER JOIN profile ON post.profile_id = profile.id
	LEFT JOIN post_image ON post.id = post_image.post_id
	INNER JOIN account ON profile.user_id = account.id
WHERE category_id = ANY($1::int[]) and is_primary = true;