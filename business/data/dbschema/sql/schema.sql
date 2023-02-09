-- Version: 1.01
-- Description: Create table users
CREATE TABLE users (
	user_id       UUID        NOT NULL,
	name          TEXT        NOT NULL,
	email         TEXT UNIQUE NOT NULL,
	roles         TEXT[]      NOT NULL,
	password_hash TEXT        NOT NULL,
    department    TEXT        NULL,
    enabled       BOOLEAN     NOT NULL,
	date_created  TIMESTAMP   NOT NULL,
	date_updated  TIMESTAMP   NOT NULL,

	PRIMARY KEY (user_id)
);

-- Version: 1.02
-- Description: Create table products
CREATE TABLE products (
	product_id   UUID      NOT NULL,
	name         TEXT      NOT NULL,
	cost         INT       NOT NULL,
	quantity     INT       NOT NULL,
	user_id      UUID      NOT NULL,
	date_created TIMESTAMP NOT NULL,
	date_updated TIMESTAMP NOT NULL,

	PRIMARY KEY (product_id),
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Version: 1.03
-- Description: Create table sales
CREATE TABLE sales (
	sale_id      UUID      NOT NULL,
	user_id      UUID      NOT NULL,
	product_id   UUID      NOT NULL,
	quantity     INT       NOT NULL,
	paid         INT       NOT NULL,
	date_created TIMESTAMP NOT NULL,

	PRIMARY KEY (sale_id),
	FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
	FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
);
