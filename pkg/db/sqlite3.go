package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const DefaultSqlitePath = "./data/app.db"

type SQLite struct {
	DB *sql.DB
}

type SQLiteConfig struct {
	Filepath string
}

func New(sc SQLiteConfig) (*SQLite, error) {
	if err := validateFilepath(sc.Filepath); err != nil {
		return nil, err
	}

	sqlDb, err := sql.Open("sqlite3", sc.Filepath)
	if err != nil {
		return nil, err
	}

	sqlDb.Exec("PRAGMA journal_mode = WAL;")
	sqlDb.Exec("PRAGMA synchronous = NORMAL;")
	sqlDb.Exec("PRAGMA foreign_keys = ON;")
	sqlDb.Exec("PRAGMA cache_size = 10000;")

	if err := sqlDb.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	return &SQLite{DB: sqlDb}, nil
}

func validateFilepath(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	return nil
}

func (s *SQLite) Ping() error {
	return s.DB.Ping()
}

func (s *SQLite) IsReady() error {
	row := s.DB.QueryRow("SELECT 1")
	var result int
	return row.Scan(&result)
}

func (s *SQLite) Close() error {
	return s.DB.Close()
}
