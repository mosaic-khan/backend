CREATE TYPE gender AS ENUM ('male', 'female', 'other', 'prefer not to say');

CREATE TABLE IF NOT EXISTS profile (
    id BIGSERIAL PRIMARY KEY ,
    user_id BIGINT REFERENCES "account"(id) not null,
    first_name VARCHAR(40) NOT NULL DEFAULT '',
    last_name VARCHAR(40) NOT NULL DEFAULT '',
    gender GENDER NOT NULL DEFAULT 'prefer not to say',
    birth_day DATE,
    profile_pic_address TEXT NOT NULL DEFAULT '',
    city_id SMALLINT REFERENCES "city"(id),
    bio varchar(140) NOT NULL DEFAULT ''
)