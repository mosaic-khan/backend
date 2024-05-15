CREATE TABLE IF NOT EXISTS post_image (
    id SERIAL PRIMARY KEY,
    post_id BIGINT REFERENCES "post"(id) ON DELETE CASCADE NOT NULL ,
    image_url VARCHAR(50) UNIQUE NOT NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX IF NOT EXISTS idx_post_image_profile_id ON post_image (post_id);