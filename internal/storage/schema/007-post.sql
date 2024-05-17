CREATE TABLE IF NOT EXISTS post (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(64) NOT NULL,
    description TEXT NOT NULL,
    category_id SMALLINT NOT NULL,
    num_images SMALLINT NOT NULL CHECK(num_images <= 10 AND num_images > 0),
    num_likes INT NOT NULL DEFAULT 0 CHECK(num_likes >= 0),
    profile_id BIGINT NOT NULL,
    FOREIGN KEY(profile_id)
        REFERENCES profile(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY(category_id)
        REFERENCES category(id)
        ON UPDATE CASCADE
        ON DELETE NO ACTION
);

CREATE EXTENSION IF NOT EXISTS pg_trgm;