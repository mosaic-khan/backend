-- Trigger for incrementing the usage number for and ingredient
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
$$

CREATE TRIGGER insert_post_ingredient
	AFTER INSERT
	ON post_has_ingredient
	FOR EACH ROW
		EXECUTE PROCEDURE inc_ingredient_usage();

 
-- Trigger for decrementing the usage number for and ingredient
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
$$


CREATE TRIGGER delete_post_ingredient
	AFTER DELETE
	ON post_has_ingredient
	FOR EACH ROW
		EXECUTE PROCEDURE dec_ingredient_usage();