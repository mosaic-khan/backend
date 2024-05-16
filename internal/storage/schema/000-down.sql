DROP TABLE IF EXISTS "like_comment";
DROP TABLE IF EXISTS "comment";

DROP TABLE IF EXISTS "profile_like_post";
DROP TABLE IF EXISTS "post_image";
DROP TABLE IF EXISTS "post_has_ingredient";
DROP TABLE IF EXISTS "ingredient";
DROP TABLE IF EXISTS "post";
DROP TABLE IF EXISTS "category";
DROP TABLE IF EXISTS "follow";
DROP TABLE IF EXISTS "profile";
DROP TABLE IF EXISTS "city";
DROP TABLE IF EXISTS "account";
DROP TABLE IF EXISTS "signup";

DROP TYPE IF EXISTS "gender";

DROP TRIGGER IF EXISTS insert_post_ingredient on post_has_ingredient;
DROP FUNCTION IF EXISTS inc_ingredient_usage;
DROP TRIGGER IF EXISTS delete_post_ingredient on post_has_ingredient;
DROP FUNCTION IF EXISTS dec_ingredient_usage;

DROP TRIGGER IF EXISTS like_post on profile_like_post;
DROP FUNCTION IF EXISTS inc_post_likes;
DROP TRIGGER IF EXISTS dislike_post on profile_like_post;
DROP FUNCTION IF EXISTS dec_post_likes;
