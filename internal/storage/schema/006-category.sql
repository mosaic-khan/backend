CREATE TABLE category (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL,
    parent SMALLINT,
    level SMALLINT,
    FOREIGN KEY(parent)
        REFERENCES category(id)
        ON UPDATE CASCADE
        ON DELETE CASCADE
);

INSERT INTO category(id, name, parent, level)
VALUES
    (1, 'فست فود', NULL, 0),
    (2, 'ایرانی', NULL, 0),
    (3, 'خورشت', 2, 1),
    (4, 'کباب', 2, 1),
    (5, 'نوشیدنی', NULL, 0),
    (6, 'نوشیدنی گرم', 5, 1),
    (7, 'نوشیدنی سرد', 5, 1),
    (8, 'نان', NULL, 0),
    (9, 'کیک و شیرینی', NULL, 0),
    (10, 'دریایی', NULL, 0),
    (11, 'بین الملل', NULL, 0),
    (12, 'صبحانه', NULL, 0),
    (13, 'سایر', NULL, 0),
    (14, 'سالاد', 13, 1),
    (15, 'ترشی', 13, 1),
    (16, 'تنقلات', 13, 1);