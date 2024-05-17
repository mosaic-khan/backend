CREATE TABLE IF NOT EXISTS profile_like_comment (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL,
    comment_id BIGINT NOT NULL,
    FOREIGN KEY(profile_id)
        REFERENCES profile(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY(comment_id)
        REFERENCES comment(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE FUNCTION inc_num_comment_likes()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS 
$$
BEGIN
    UPDATE comment
    SET num_likes = num_likes + 1
    WHERE id = NEW.comment_id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER insert_comment_like
    AFTER INSERT
    ON profile_like_comment
    FOR EACH ROW
        EXECUTE PROCEDURE inc_num_comment_likes();


CREATE FUNCTION dec_num_comment_likes()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS 
$$
BEGIN
    UPDATE comment
    SET num_likes = num_likes - 1
    WHERE id = OLD.comment_id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER delete_comment_like
    AFTER DELETE
    ON profile_like_comment
    FOR EACH ROW
        EXECUTE PROCEDURE dec_num_comment_likes();