-- name: InsertNewUser :exec
INSERT INTO account (email, username, password)
VALUES ($1, $2, $3);

-- name: UpdateUserInfo :exec
UPDATE account
SET	first_name = $1,
	last_name  = $2,
	gender	   = $3,
	birth_day  = $4
WHERE username = $5;

-- name: GetPasswordByEmail :one
SELECT password
FROM account
WHERE email = $1;

-- name: GetPasswordByUsername :one
select password
FROM account
WHERE username = $1;

-- name: ExistsEmail :one
SELECT count(*)
FROM account
WHERE email = $1;

-- name: ExistsUsername :one
SELECT count(*)
FROM account
WHERE username = $1;