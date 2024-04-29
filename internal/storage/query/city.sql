-- name: GetCities :many
SELECT *
FROM city
WHERE name LIKE '%' || $1 || '%'
LIMIT 20;


