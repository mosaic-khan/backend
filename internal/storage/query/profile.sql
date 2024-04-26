

-- name: UpdateProfileInfo :exec
UPDATE profile
SET	first_name = $1,
        last_name  = $2,
        gender	   = $3,
        birth_day  = $4,
        bio = $5,
        city_id = $6
WHERE user_id = $7;
