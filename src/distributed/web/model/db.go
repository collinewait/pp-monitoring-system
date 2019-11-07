package model

import "database/sql"

var db *sql.DB

func init() {
	var err error
	db, _ = sql.Open("postgres",
		"postgres://collinewaitire:wait@localhost/distributed?sslmode=disable")

	if err != nil {
		panic(err.Error())
	}
}
