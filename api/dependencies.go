package api

import (
	"log"
	"oapi-to-rest/pkg/db"
	"oapi-to-rest/pkg/env"
	"oapi-to-rest/pkg/errlib"
	"oapi-to-rest/pkg/jwt"

	"github.com/jmoiron/sqlx"
)

type Dependencies struct {
	DbSqlite *db.SQLite
	Sqlx     *sqlx.DB

	ErrorHandler *errlib.ErrorHandler
	Jwt          *jwt.TokenManager
}

func InitDependencies(cfg *env.Config) Dependencies {

	var dep Dependencies

	// errorHandler
	dep.ErrorHandler = errlib.NewErrorHandler(cfg.DebugMode)

	tm, err := jwt.NewRSAJwtInit(&cfg.Jwt)
	if err != nil {
		log.Println("error init rsa jwt, err: %w", err)
		tm = nil
	}

	dep.Jwt = tm

	// db
	if cfg.InitSqlite {
		dbcfg := db.SQLiteConfig{
			Filepath: db.DefaultSqlitePath,
		}

		// connect sqlite
		sqlite, err := db.New(dbcfg)
		if err != nil {
			log.Fatalf("error connect sqlite: %v", err)
		}

		dep.DbSqlite = sqlite

		err = dep.DbSqlite.Ping()
		if err != nil {
			log.Fatalf("error ping sqlite: %v", err)
		}
		err = dep.DbSqlite.IsReady()
		if err != nil {
			log.Fatalf("error sqlite not ready: %v", err)
		}

		// sqlx deps wrap sql.DB with sqlx
		dep.Sqlx = sqlx.NewDb(dep.DbSqlite.DB, "sqlite3")
	}

	return dep
}
