package main

import (
	"database/sql"
	"flag"
	"log"

	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed default_sqlite_ddl.sql
var defaultDDL []byte

func main() {
	dbPath := flag.String("db", "/data/app.db", "path to sqlite db")
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("open DB: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(string(defaultDDL)); err != nil {
		log.Fatalf("create default table failed: %v", err)
	}

	log.Println("default sqlite initialized.")
}
