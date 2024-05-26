-- name: InsertPost :one
INSERT INTO post (title, description, category_id, num_images, profile_id)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetPost :one
SELECT post.id, post.title, post.description, category.name as category, account.username, profile.profile_pic_address, post.num_images, post.num_likes, post.num_comments, post.profile_id,
    (SELECT EXISTS
        (SELECT
        FROM profile_pin_post
        WHERE profile_pin_post.profile_id = $2 AND profile_pin_post.post_id = $1)
    ) AS pinned,
    (SELECT EXISTS
        (SELECT
        FROM profile_like_post
        WHERE profile_like_post.profile_id = $2 AND profile_like_post.post_id = $1)
    ) AS liked
FROM post
         JOIN profile on profile.id = post.profile_id
         JOIN account on account.id = profile.user_id
         JOIN category on post.category_id = category.id
WHERE post.id = $1;

-- name: GetPostProfileId :one
SELECT profile_id
FROM post
WHERE id = $1;

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
SELECT post.id, post.title, post.description, post_image.image_url, post.num_likes, post.num_comments,
       (SELECT EXISTS
                   (SELECT 1 FROM profile_like_post
                    WHERE profile_like_post.profile_id = $2 AND post_id = post.id)
       ) AS isLiked
FROM post
    LEFT JOIN post_image on post.id = post_image.post_id
WHERE post.profile_id = $1 and post_image.is_primary = true
ORDER BY post.id DESC
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
WHERE category_id = ANY($1::int[]) and is_primary = true
ORDER BY post.num_likes
OFFSET $2
LIMIT 20;


-- name: SearchName :many
SELECT post.id, post.title, post.description, post_image.image_url,
       profile.profile_pic_address, account.username
FROM post
    JOIN profile on post.profile_id = profile.id
    JOIN account on profile.user_id = account.id
    LEFT JOIN post_image on post.id = post_image.post_id
WHERE post_image.is_primary = true and similarity(post.title, sqlc.arg(name)) > 0.5
ORDER BY (similarity(post.title, sqlc.arg(name)), post.num_likes) DESC
OFFSET sqlc.arg(page)
LIMIT 20;


-- name: SearchIngredient :many
WITH selected_post_id AS (
      SELECT post_has_ingredient.post_id AS p_id
      FROM post_has_ingredient
               JOIN ingredient on post_has_ingredient.ingredient_id = ingredient.id
      WHERE (sqlc.arg(includeCnt) = 0 OR ingredient.name = ANY (sqlc.arg(include)::TEXT[]))
      GROUP BY post_has_ingredient.post_id
      HAVING COUNT(DISTINCT ingredient.id) >= sqlc.arg(includeCnt)
      INTERSECT
      SELECT post_has_ingredient.post_id AS p_id
      FROM post_has_ingredient
               JOIN ingredient on post_has_ingredient.ingredient_id = ingredient.id
      WHERE NOT EXISTS (SELECT 1
                        FROM post_has_ingredient phi
                            JOIN ingredient ing on phi.ingredient_id = ing.id
                        WHERE phi.post_id = post_has_ingredient.post_id
                            and ing.name = ANY (sqlc.arg(exclude)::TEXT[]))
)
SELECT post.id, post.title, post.description,
       post_image.image_url, account.username, profile.profile_pic_address
FROM selected_post_id
    JOIN post on post.id = selected_post_id.p_id
    JOIN profile on post.profile_id = profile.id
    JOIN account on profile.user_id = account.id
    LEFT JOIN post_image on post.id = post_image.post_id
WHERE post_image.is_primary = true
ORDER BY post.num_likes DESC
OFFSET sqlc.arg(page)
LIMIT 20;


-- name: MixedSearch :many
WITH selected_post_id AS (
    SELECT post_has_ingredient.post_id AS p_id
    FROM post_has_ingredient
             JOIN ingredient on post_has_ingredient.ingredient_id = ingredient.id
    WHERE (sqlc.arg(includeCnt) = 0 OR ingredient.name = ANY (sqlc.arg(include)::TEXT[]))
    GROUP BY post_has_ingredient.post_id
    HAVING COUNT(DISTINCT ingredient.id) >= sqlc.arg(includeCnt)
    INTERSECT
    SELECT post_has_ingredient.post_id AS p_id
    FROM post_has_ingredient
             JOIN ingredient on post_has_ingredient.ingredient_id = ingredient.id
    WHERE NOT EXISTS (SELECT 1
                      FROM post_has_ingredient phi
                               JOIN ingredient ing on phi.ingredient_id = ing.id
                      WHERE phi.post_id = post_has_ingredient.post_id
                        and ing.name = ANY (sqlc.arg(exclude)::TEXT[]))
    INTERSECT
    SELECT post.id
    FROM post
    WHERE category_id = ANY(sqlc.arg(categories)::int[]) OR array_length(sqlc.arg(categories)::int[], 1) = 0
    INTERSECT
    SELECT post.id
    FROM post
    WHERE post.title = '' OR similarity(post.title, sqlc.arg(name)) > 0.5
)
SELECT post.id, post.title, post.description,
       post_image.image_url, account.username, profile.profile_pic_address
FROM selected_post_id
         JOIN post on post.id = selected_post_id.p_id
         JOIN profile on post.profile_id = profile.id
         JOIN account on profile.user_id = account.id
         LEFT JOIN post_image on post.id = post_image.post_id
WHERE post_image.is_primary = true
ORDER BY post.num_likes DESC
OFFSET sqlc.arg(page)
    LIMIT 20;
