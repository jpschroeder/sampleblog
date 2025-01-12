package main

import (
	"database/sql"
)

var _db *sql.DB

func initDB(dbFile string) (*sql.DB, error) {
	var err error
	_db, err = sql.Open("sqlite3", dbFile)
	return _db, err
}

func getDB() *sql.DB {
	return _db
}
