CREATE TABLE profile_like_post (
	post_id BIGINT,
	profile_id BIGINT,
	PRIMARY KEY (post_id, profile_id),
	FOREIGN KEY (post_id)
		REFERENCES post(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE,
	FOREIGN KEY (profile_id)
		REFERENCES profile(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE
);


-- related triggers
CREATE FUNCTION inc_post_likes()
   RETURNS TRIGGER 
   LANGUAGE PLPGSQL
AS 
$$
BEGIN
	UPDATE post
	SET num_likes = num_likes + 1
	WHERE id = NEW.post_id;
	RETURN NEW;
END;
$$;

CREATE TRIGGER like_post
	AFTER INSERT
	ON profile_like_post
	FOR EACH ROW
		EXECUTE PROCEDURE inc_post_likes();



CREATE FUNCTION dec_post_likes()
   RETURNS TRIGGER 
   LANGUAGE PLPGSQL
AS 
$$
BEGIN
	UPDATE post
	SET num_likes = num_likes - 1
	WHERE id = OLD.post_id;
	RETURN NEW;
END;
$$;


CREATE TRIGGER dislike_post
	AFTER DELETE
	ON profile_like_post
	FOR EACH ROW
		EXECUTE PROCEDURE dec_post_likes();