CREATE TABLE comment (
    id BIGSERIAL PRIMARY KEY,
    parent_id BIGINT REFERENCES comment(id),
    post_id BIGINT NOT NULL REFERENCES post(id),
    profile_id BIGINT NOT NULL REFERENCES profile(id),
    comment TEXT NOT NULL,
    time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIME
);

CREATE TABLE like_comment (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES profile(id),
    comment_id BIGINT NOT NULL REFERENCES comment(id)
);