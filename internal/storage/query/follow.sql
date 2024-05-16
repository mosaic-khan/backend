-- name: Follow :exec
INSERT INTO follow (follower, following) VALUES ($1, $2);

-- name: Unfollow :exec
DELETE FROM follow
WHERE follower = $1 and following = $2;

-- name: GetFollowingCnt :one
SELECT count(*)
FROM follow
WHERE follower = $1;

-- name: GetFollowerCnt :one
SELECT count(*)
FROM follow
WHERE following = $1;

-- name: FollowStatus :one
SELECT EXISTS (
    SELECT 1
    FROM follow
    WHERE follower = $1 AND following = $2
);


-- name: FollowingList :many
SELECT profile.id, account.username,  profile.first_name || ' ' || profile.last_name, profile.profile_pic_address
FROM follow JOIN profile on follow.following = profile.id
JOIN account on profile.user_id = account.id
WHERE follower = $1;

-- name: FollowerList :many
SELECT profile.id, account.username,  profile.first_name || ' ' || profile.last_name, profile.profile_pic_address
FROM follow join profile on follow.follower = profile.id
JOIN account on profile.user_id = account.id
WHERE following = $1;