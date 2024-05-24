CREATE TABLE profile_pin_post (
    profile_id BIGINT NOT NULL,
    post_id BIGINT NOT NULL,
    PRIMARY KEY(profile_id, post_id),
    FOREIGN KEY(profile_id)
        REFERENCES profile(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    FOREIGN KEY(post_id)
        REFERENCES post(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);