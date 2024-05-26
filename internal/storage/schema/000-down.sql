DROP TRIGGER IF EXISTS insert_report on report_comment;
DROP FUNCTION IF EXISTS inc_num_report;

DROP TRIGGER IF EXISTS insert_post_ingredient on post_has_ingredient;
DROP FUNCTION IF EXISTS inc_ingredient_usage;
DROP TRIGGER IF EXISTS delete_post_ingredient on post_has_ingredient;
DROP FUNCTION IF EXISTS dec_ingredient_usage;

DROP TRIGGER IF EXISTS like_post on profile_like_post;
DROP FUNCTION IF EXISTS inc_post_likes;
DROP TRIGGER IF EXISTS dislike_post on profile_like_post;
DROP FUNCTION IF EXISTS dec_post_likes;

DROP TRIGGER IF EXISTS insert_comment on comment;
DROP FUNCTION IF EXISTS inc_num_comments;
DROP TRIGGER IF EXISTS delete_comment on comment;
DROP FUNCTION IF EXISTS dec_num_comments;

DROP TRIGGER IF EXISTS insert_comment_like on profile_like_comment;
DROP FUNCTION IF EXISTS inc_num_comment_likes;
DROP TRIGGER IF EXISTS delete_comment_like on profile_like_comment;
DROP FUNCTION IF EXISTS dec_num_comment_likes;

DROP TABLE IF EXISTS "report" CASCADE ;
DROP TABLE IF EXISTS "profile_like_post" CASCADE ;
DROP TABLE IF EXISTS "profile_like_comment" CASCADE ;
DROP TABLE IF EXISTS "comment" CASCADE ;
DROP TABLE IF EXISTS "post_image" CASCADE ;
DROP TABLE IF EXISTS "post_has_ingredient" CASCADE ;
DROP TABLE IF EXISTS "ingredient" CASCADE ;
DROP TABLE IF EXISTS "post" CASCADE ;
DROP TABLE IF EXISTS "category" CASCADE ;
DROP TABLE IF EXISTS "follow" CASCADE ;
DROP TABLE IF EXISTS "profile" CASCADE ;
DROP TABLE IF EXISTS "city" CASCADE ;
DROP TABLE IF EXISTS "account" CASCADE ;
DROP TABLE IF EXISTS "signup" CASCADE ;

DROP TYPE IF EXISTS "gender";
