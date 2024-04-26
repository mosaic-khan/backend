-- name: GetCities :many
SELECT *
FROM city
where name like $1
LIMIT 20;


