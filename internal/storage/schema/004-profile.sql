CREATE TYPE gender AS ENUM ('male', 'female', 'other', 'prefer not to say');

CREATE TABLE IF NOT EXISTS profile (
    id BIGSERIAL PRIMARY KEY ,
    user_id BIGINT REFERENCES "account"(id),
    first_name VARCHAR(40),
    last_name VARCHAR(40),
    gender GENDER NOT NULL DEFAULT 'prefer not to say',
    birth_day DATE,
    profile_pic_address TEXT,
    city_id SMALLINT REFERENCES "city"(id),
    bio varchar(140)
)