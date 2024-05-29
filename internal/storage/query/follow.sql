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
SELECT profile.id, account.username,  (profile.first_name || ' ' || profile.last_name) AS name, profile.profile_pic_address,
       (SELECT EXISTS(SELECT 1 FROM follow f WHERE f.follower = sqlc.arg(myProfile) and f.following = sqlc.arg(profileID))) AS is_followed
FROM follow JOIN profile on follow.following = profile.id
    JOIN account on profile.user_id = account.id
WHERE follower = sqlc.arg(profileID);

-- name: FollowerList :many
SELECT profile.id, account.username,  (profile.first_name || ' ' || profile.last_name) AS name, profile.profile_pic_address,
       (SELECT EXISTS(SELECT 1 FROM follow f WHERE f.follower = sqlc.arg(myProfile) and f.following = sqlc.arg(profileID))) AS is_followed
FROM follow JOIN profile on follow.following = profile.id
    JOIN account on profile.user_id = account.id
WHERE following = sqlc.arg(profileID);