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

Example Usage:

	package main

	import (
		"database/sql"
		"log"

		"github.com/ardanlabs/darwin"
		_ "github.com/go-sql-driver/mysql"
	)

	var (
		migrations = []darwin.Migration{
			{
				Version:     1,
				Description: "Creating table posts",
				Script: `CREATE TABLE posts (
							id INT 		auto_increment,
							title 		VARCHAR(255),
							PRIMARY KEY (id)
						) ENGINE=InnoDB CHARACTER SET=utf8;`,
			},
			{
				Version:     2,
				Description: "Adding column body",
				Script:      "ALTER TABLE posts ADD body TEXT AFTER title;",
			},
		}
	)

	func main() {
		database, err := sql.Open("mysql", "root:@/darwin")
		if err != nil {
			log.Fatal(err)
		}

		driver, err := darwin.NewGenericDriver(database, darwin.MySQLDialect{})
		if err != nil {
			log.Fatal(err)
		}

		d := darwin.New(driver, migrations)
		if err := d.Migrate(); err != nil {
			log.Println(err)
		}
	}
*/
package darwin
