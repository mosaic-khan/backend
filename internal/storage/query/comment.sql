-- name: AddComment :one
INSERT INTO comment(post_id, profile_id, comment)
VALUES ($1, $2, $3)
RETURNING id;

-- name: AddReply :one
WITH parent AS (
    SELECT post_id
    FROM comment
    WHERE id = $1
)
INSERT INTO comment(parent_id, post_id, profile_id, comment)
SELECT $1, parent.post_id, $2, $3
FROM parent
RETURNING id;


-- name: LikeCommentOrReply :exec
INSERT INTO profile_like_comment(profile_id, comment_id)
VALUES ($1, $2);

-- name: DislikeCommentOrReply :exec
DELETE
FROM profile_like_comment
WHERE profile_id = $1 and comment_id = $2;


-- name: GetPostsComments :many
SELECT comment.id, account.username, profile.first_name, profile.profile_pic_address, comment.comment, comment.num_likes,
       (SELECT EXISTS
           (SELECT 1 FROM profile_like_comment
            WHERE profile_like_comment.comment_id = comment.id and profile_like_comment.profile_id = $1)
       ) AS isLiked,
       (SELECT EXISTS
           (SELECT 1 FROM comment c
                WHERE c.parent_id = comment.id
           )
       ) AS has_replies,
       comment.profile_id = $1 AS owned,
       comment.time
FROM comment
    JOIN profile on comment.profile_id = profile.id
    JOIN account on profile.user_id = account.id
WHERE comment.parent_id IS NULL and comment.post_id = $2 and num_report < 1000;


-- name: GetReplies :many
WITH RECURSIVE replies AS (
SELECT comment.id, account.username, profile.first_name, profile.profile_pic_address, comment.comment, comment.num_likes,
       (lk.isLiked IS NOT NULL) AS isLiked,
       (SELECT EXISTS
                   (SELECT 1 FROM comment c
                    WHERE c.parent_id = comment.id
                   )
       ) AS has_replies,
       comment.time,
       comment.parent_id,
       comment.profile_id = $1 AS owned
FROM comment
         JOIN profile on comment.profile_id = profile.id
         JOIN account on profile.user_id = account.id
         LEFT JOIN (
            SELECT comment_id AS cmnt_id, 1 AS isLiked
            FROM profile_like_comment
            WHERE profile_like_comment.profile_id = $1
         ) AS lk on comment.id = lk.cmnt_id
WHERE comment.parent_id = $2 AND num_report < 1000
UNION
SELECT comment.id, account.username, profile.first_name, profile.profile_pic_address, comment.comment, comment.num_likes,
       (lk.isLiked IS NOT NULL) AS isLiked,
       (SELECT EXISTS
                   (SELECT 1 FROM comment c
                    WHERE c.parent_id = comment.id
                   )
       ) AS has_replies,
       comment.time,
       comment.parent_id,
       comment.profile_id = $1 AS owned
FROM comment
         JOIN profile on comment.profile_id = profile.id
         JOIN account on profile.user_id = account.id
         LEFT JOIN (
            SELECT comment_id AS cmnt_id, 1 AS isLiked
            FROM profile_like_comment
            WHERE profile_like_comment.profile_id = $1
         ) AS lk on comment.id = lk.cmnt_id
         JOIN replies ON comment.parent_id = replies.id
WHERE num_report < 1000
)
SELECT *
FROM replies;



-- name: ReportComment :exec
INSERT INTO report_comment (profile_id, comment_id) values ($1, $2);



-- name: GetComment :one
SELECT comment.id, account.username, profile.first_name, profile.profile_pic_address, comment.comment, comment.num_likes,
       (SELECT EXISTS
           (SELECT 1 FROM profile_like_comment
            WHERE profile_like_comment.comment_id = comment.id and profile_like_comment.profile_id = $1)
       ) AS isLiked,
       (SELECT EXISTS
           (SELECT 1 FROM comment c
                WHERE c.parent_id = comment.id
           )
       ) AS has_replies,
       comment.profile_id = $1 AS owned,
       comment.time
FROM comment
    JOIN profile on comment.profile_id = profile.id
    JOIN account on profile.user_id = account.id
WHERE comment.id = $2;

-- name: DeleteComment :exec
DELETE
FROM comment
WHERE profile_id = $1 AND id = $2;
