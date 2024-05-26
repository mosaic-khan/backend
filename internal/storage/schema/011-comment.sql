CREATE TABLE IF NOT EXISTS comment (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT,
    post_id BIGINT NOT NULL REFERENCES post(id),
    profile_id BIGINT NOT NULL,
    num_likes INT NOT NULL DEFAULT 0 CHECK(num_likes >= 0),
    comment TEXT NOT NULL,
    time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    num_report INT NOT NULL DEFAULT 0,
    FOREIGN KEY(parent_id)
        REFERENCES comment(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY(profile_id)
        REFERENCES profile(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY(post_id)
        REFERENCES post(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

-- related triggers
CREATE FUNCTION inc_num_comments()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS 
$$
BEGIN
    UPDATE post
    SET num_comments = num_comments + 1
    WHERE id = NEW.post_id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER insert_comment
    AFTER INSERT
    ON comment
    FOR EACH ROW
        EXECUTE PROCEDURE inc_num_comments();


CREATE FUNCTION dec_num_comments()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS 
$$
BEGIN
    UPDATE post
    SET num_comments = num_comments - 1
    WHERE id = OLD.post_id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER delete_comment
    AFTER DELETE
    ON comment
    FOR EACH ROW
        EXECUTE PROCEDURE dec_num_comments();