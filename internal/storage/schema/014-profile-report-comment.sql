CREATE TABLE IF NOT EXISTS report_comment (
    id SERIAL NOT NULL PRIMARY KEY,
    profile_id  BIGINT REFERENCES profile(id) ON DELETE CASCADE NOT NULL,
    comment_id BIGINT REFERENCES comment(id) ON DELETE CASCADE NOT NULL,
    UNIQUE (profile_id, comment_id)
);


-- related triggers
CREATE FUNCTION inc_num_report()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    UPDATE comment
    SET num_report = num_report + 1
    WHERE id = NEW.comment_id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER insert_report
    AFTER INSERT
    ON report_comment
    FOR EACH ROW
EXECUTE PROCEDURE inc_num_comments();
