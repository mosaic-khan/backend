-- name: AddComment :exec
INSERT INTO comment(post_id, profile_id, comment)
VALUES ($1, $2, $3);

-- name: AddReply :exec
INSERT INTO comment(parent_id, post_id, profile_id, comment)
VALUES ($1, $2, $3, $4);


-- LikeCommentOrReply :exec
INSERT INTO like_comment(profile_id, comment_id)
VALUES ($1, $2);

-- name: GetPostsComments :many
SELECT account.username, profile.profile_pic_address, comment.comment,
       (SELECT EXISTS
           (SELECT 1 FROM like_comment
            WHERE like_comment.comment_id = comment.id and like_comment.profile_id = $1)
       ) AS isLiked,
       (SELECT EXISTS
           (SELECT 1 FROM comment c
                WHERE c.parent_id = comment.id
           )
       ) AS has_replies,
       comment.time
FROM comment
    JOIN profile on comment.profile_id = profile.id
    JOIN account on profile.user_id = account.id
WHERE comment.parent_id IS NULL and comment.post_id = $2;


-- name: GetReplies :many
SELECT account.username, profile.profile_pic_address, comment.comment,
       (lk.isLiked IS NOT NULL) AS isLiked,
       (SELECT EXISTS
                   (SELECT 1 FROM comment c
                    WHERE c.parent_id = comment.id
                   )
       ) AS has_replies,
       comment.time
FROM comment
         JOIN profile on comment.profile_id = profile.id
         JOIN account on profile.user_id = account.id
         LEFT JOIN (
            SELECT comment_id AS cmnt_id, 1 AS isLiked
            FROM like_comment
            WHERE like_comment.profile_id = $1
         ) AS lk on comment.id = lk.cmnt_id
WHERE comment.parent_id = $3 and comment.post_id = $2;
