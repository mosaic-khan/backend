-- name: CreateProfile :one
INSERT INTO profile (user_id)
VALUES ($1)
RETURNING id;

-- name: GetProfileUserID :one
SELECT user_id
FROM profile
WHERE id = $1;

-- name: GetProfileID :one
SELECT id
FROM profile
WHERE user_id = $1;

-- name: GetUserInfo :one
SELECT account.username, profile.first_name, profile.last_name,
       profile.gender, profile.birth_day, profile.bio,
       city.name AS city_name, profile.city_id, account.email, profile.profile_pic_address
FROM profile JOIN account on account.id = profile.user_id
    LEFT JOIN city on city.id = profile.city_id
WHERE profile.id = $1;


-- name: UpdateProfileInfo :exec
UPDATE profile
SET	    first_name = $1,
        last_name  = $2,
        gender	   = $3,
        birth_day  = $4,
        bio        = $5,
        city_id    = $6
WHERE id = $7;


-- name: GetProfileByUsername :one
SELECT profile.id, account.username, (profile.first_name || profile.last_name) AS name,
       profile.bio, profile.gender, city.name AS city_name, profile.profile_pic_address
FROM profile JOIN account on account.id = profile.user_id
    LEFT JOIN city on city.id = profile.city_id
WHERE account.username = $1;


-- name: GetProfileByUserID :one
SELECT profile.id, account.username, (profile.first_name || profile.last_name) AS name,
       profile.bio, profile.gender, city.name AS city_name, profile.profile_pic_address
FROM profile JOIN account on account.id = profile.user_id
             LEFT JOIN city on city.id = profile.city_id
WHERE profile.id = $1;

-- name: ChangeProfilePic :exec
UPDATE profile
SET profile_pic_address = $1
WHERE user_id = $2;

