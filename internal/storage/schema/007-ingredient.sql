CREATE TABLE IF NOT EXISTS ingredient (
	id SERIAL PRIMARY KEY,
	name VARCHAR(32) NOT NULL UNIQUE,	
	usage INT NOT NULL DEFAULT 0 CHECK(usage >= 0) 
);

INSERT INTO ingredient (name)
VALUES ('سیب'), ('سیر'), ('سویا'), ('سیب ترش'), ('سیب سفید'),
       ('سوسیس'), ('سویا سبز'), ('سویا سیاه'), ('سیب زمینی'),
       ('سماق'), ('سمنو');
