CREATE TABLE IF NOT EXISTS post (
	id BIGSERIAL PRIMARY KEY,
	title VARCHAR(64) NOT NULL,
	description TEXT NOT NULL,
	num_images SMALLINT NOT NULL CHECK(num_images <= 10 AND num_images > 0)
);