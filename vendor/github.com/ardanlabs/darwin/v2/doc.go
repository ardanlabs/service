/*
Package darwin provides a database schema evolution api for Go. The purpose of
this library is just be a library. You can implement your own way of building
the migration list. It is not recommended to put more than one database change
per migration, if some migration fail, you exactly what statement caused the
error. Also only postgres correctly handle rollback in DDL transactions.

The best way to version your migrations is like this: 1.0, 1.1, 1.2

Please read the following posts for more information on the design principles
if this package.

https://flywaydb.org/documentation/faq#downgrade
https://flywaydb.org/documentation/faq#rollback
https://flywaydb.org/documentation/faq#hot-fixes

Given this file:
	-- Version: 1.1
	-- Description: Create table users
	CREATE TABLE users (
		user_id       UUID,
		name          TEXT,
		email         TEXT UNIQUE,
		roles         TEXT[],
		password_hash TEXT,
		date_created  TIMESTAMP,
		date_updated  TIMESTAMP,

		PRIMARY KEY (user_id)
	);

	-- Version: 1.2
	-- Description: Create table products
	CREATE TABLE products (
		product_id   UUID,
		name         TEXT,
		cost         INT,
		quantity     INT,
		date_created TIMESTAMP,
		date_updated TIMESTAMP,

		PRIMARY KEY (product_id)
	);

You can write this code:

	package main

	import (
		"database/sql"
		"log"

		"github.com/ardanlabs/darwin"
		_ "github.com/go-sql-driver/mysql"
	)

	var (
		//go:embed sql/schema.sql
		schemaDoc string
	)

	func main() {
		database, err := sql.Open("mysql", "root:@/darwin")
		if err != nil {
			log.Fatal(err)
		}

		driver, err := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})
		if err != nil {
			return err
		}

		d := darwin.New(driver, darwin.ParseMigrations(schemaDoc))
		if err := d.Migrate(); err != nil {
			log.Println(err)
		}
	}
*/
package darwin
