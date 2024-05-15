CREATE TABLE IF NOT EXISTS post_has_ingredient (
    id BIGSERIAL PRIMARY KEY,
    amount VARCHAR(64),
    post_id BIGINT NOT NULL REFERENCES post(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    ingredient_id INT NOT NULL REFERENCES ingredient(id)
        ON DELETE NO ACTION
        ON UPDATE CASCADE
);


-- related triggers
CREATE FUNCTION inc_ingredient_usage()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS 
$$
BEGIN
    UPDATE ingredient
    SET usage = usage + 1
    WHERE id = NEW.ingredient_id;
	RETURN NEW;
END;
$$;

CREATE TRIGGER insert_post_ingredient
    AFTER INSERT
    ON post_has_ingredient
    FOR EACH ROW
        EXECUTE PROCEDURE inc_ingredient_usage();


CREATE FUNCTION dec_ingredient_usage()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS 
$$
BEGIN
    UPDATE ingredient
    SET usage = usage - 1
    WHERE id = OLD.ingredient_id;
    RETURN NEW;
END;
$$;


CREATE TRIGGER delete_post_ingredient
    AFTER DELETE
    ON post_has_ingredient
    FOR EACH ROW
        EXECUTE PROCEDURE dec_ingredient_usage();
