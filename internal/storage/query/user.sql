-- name: InsertUser :one
INSERT INTO account (email, username, password)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetUserPasswordByEmail :one
SELECT password
FROM account
WHERE email = $1;

-- name: GetUserPasswordByUsername :one
select password
FROM account
WHERE username = $1;

-- name: GetUserByEmail :one
SELECT *
FROM account
WHERE email = $1;

-- name: GetUserByUsername :one
SELECT *
FROM account
WHERE username = $1;

-- name: GetUserByID :one
SELECT *
FROM account
WHERE id = $1;

-- name: ExistsUserEmail :one
SELECT count(*)
FROM account
WHERE email = $1;

-- name: ExistsUserUsername :one
SELECT count(*)
FROM account
WHERE username = $1;

-- name: ResetPassword :exec
UPDATE account
SET password = $2
WHERE id = $1;


-- name: GetUserIDbyProfileID :one
SELECT account.id
FROM account
    JOIN profile on account.id = profile.user_id
WHERE profile.id = $1;

-- name: DeleteUser :exec
DELETE
FROM account
WHERE id = $1;

-- name: SearchUsername :many
WITH username_similarity AS (
    SELECT profile.id, username, first_name, profile_pic_address, similarity(username, sqlc.arg(username)) AS similarity
    FROM account INNER JOIN profile on account.id = profile.user_id
)
SELECT id, username, first_name, profile_pic_address
FROM username_similarity
WHERE similarity > 0.3 
ORDER BY similarity DESC
LIMIT 50;