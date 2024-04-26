-- name: CreateProfile :exec
INSERT INTO profile (user_id)
VALUES ($1);

-- name: GetUserInfo :one
SELECT account.username, profile.first_name, profile.last_name, profile.gender, profile.birth_day, profile.bio, city.name, account.email
FROM profile JOIN account on account.id = profile.user_id
    JOIN city on city.id = profile.city_id
WHERE user_id = $1;


-- name: UpdateProfileInfo :exec
UPDATE profile
SET	    first_name = $1,
        last_name  = $2,
        gender	   = $3,
        birth_day  = $4,
        bio        = $5,
        city_id    = $6
WHERE user_id = $7;


-- name: GetProfileByUsername :one
SELECT profile.id, account.username, profile.first_name || profile.last_name,
       profile.bio, profile.gender, city.name, profile.profile_pic_address
FROM profile JOIN account on account.id = profile.user_id
    JOIN city on city.id = profile.city_id
WHERE account.username = $1;


-- name: GetProfileByUserID :one
SELECT profile.id, account.username, profile.first_name || profile.last_name,
       profile.bio, profile.gender, city.name, profile.profile_pic_address
FROM profile JOIN account on account.id = profile.user_id
             JOIN city on city.id = profile.city_id
WHERE account.id = $1;

-- name: ChangeProfilePic :exec
UPDATE profile
SET profile_pic_address = $1
WHERE user_id = $2;

