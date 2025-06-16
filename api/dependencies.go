package api

import (
	"log"
	"oapi-to-rest/pkg/db"
	"oapi-to-rest/pkg/env"
	"oapi-to-rest/pkg/errlib"
)

type Dependencies struct {
	DbSqlite     *db.SQLite
	ErrorHandler *errlib.ErrorHandler
}

func InitDependencies(cfg *env.Config) Dependencies {

	var dep Dependencies

	// errorHandler
	dep.ErrorHandler = errlib.NewErrorHandler(cfg.DebugMode)

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
	}

	return dep
}
