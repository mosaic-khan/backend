CREATE TYPE gender AS ENUM ('male', 'female', 'other', 'prefer not to say');

CREATE TABLE IF NOT EXISTS profile (
    id BIGSERIAL PRIMARY KEY ,
    user_id BIGINT NOT NULL,
    first_name VARCHAR(40) NOT NULL DEFAULT '',
    last_name VARCHAR(40) NOT NULL DEFAULT '',
    gender GENDER NOT NULL DEFAULT 'prefer not to say',
    birth_day DATE,
    profile_pic_address TEXT NOT NULL DEFAULT '',
    city_id SMALLINT,
    bio varchar(140) NOT NULL DEFAULT '',
    FOREIGN KEY(user_id)
        REFERENCES account(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY(city_id)
        REFERENCES city(id)
);


-- related trigger
-- delete all profile's posts
CREATE FUNCTION delete_profile_post()
   RETURNS TRIGGER 
   LANGUAGE PLPGSQL
AS 
$$
BEGIN
    DELETE
    FROM post
    WHERE id IN (
		SELECT post_id
		FROM profile_has_post
		WHERE profile_id = OLD.id
	);
	RETURN OLD;
END;
$$;

CREATE TRIGGER delete_profile
	BEFORE DELETE
	ON profile
	FOR EACH ROW
		EXECUTE PROCEDURE delete_profile_post();