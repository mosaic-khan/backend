-- name: InsertSignin :one
INSERT INTO signin (email, username, password, verification_code, expire)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: GetSigninCode :one
SELECT verification_code
FROM signin
WHERE id = $1;

-- name: DeleteSignin :exec
DELETE
FROM signin
WHERE id = $1;