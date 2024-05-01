-- name: GetCities :many
SELECT *
FROM city
WHERE name LIKE '%' || sqlc.arg(name) || '%'
LIMIT 20;


