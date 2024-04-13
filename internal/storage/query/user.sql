-- name: InsertUser :exec
INSERT INTO account (email, username, password)
VALUES ($1, $2, $3);

-- name: UpdateUserInfo :exec
UPDATE account
SET	first_name = $1,
	last_name  = $2,
	gender	   = $3,
	birth_day  = $4
WHERE id = $5;

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

-- name: ExistsUserEmail :one
SELECT count(*)
FROM account
WHERE email = $1;

-- name: ExistsUserUsername :one
SELECT count(*)
FROM account
WHERE username = $1;