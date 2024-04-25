CREATE TABLE IF NOT EXISTS follow (
    follower BIGINT REFERENCES "profile"(id) NOT NULL,
    following BIGINT REFERENCES "profile"(id) NOT NULL,
    UNIQUE (follower, following)
)