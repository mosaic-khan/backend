CREATE TYPE gender AS ENUM ('male', 'female', 'other', 'prefer not to say');


CREATE TABLE account (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(571) NOT NULL UNIQUE,
    username VARCHAR(32) NOT NULL UNIQUE,
    first_name VARCHAR(40),
    last_name VARCHAR(40),
    gender GENDER,
    birth_day DATE,
    creation_date DATE  NOT NULL DEFAULT CURRENT_DATE,
    password CHAR(60) NOT NULL
);