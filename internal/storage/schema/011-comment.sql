CREATE TABLE IF NOT EXISTS comment (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES comment(id) ON DELETE CASCADE,
    post_id BIGINT NOT NULL REFERENCES post(id),
    profile_id BIGINT NOT NULL REFERENCES profile(id),
    comment TEXT NOT NULL,
    time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS like_comment (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES profile(id),
    comment_id BIGINT NOT NULL REFERENCES comment(id),
    UNIQUE (profile_id, comment_id)
);