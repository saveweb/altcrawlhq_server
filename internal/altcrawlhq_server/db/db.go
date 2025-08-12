package db

import (
	"database/sql"
	_ "embed"
	"log/slog"
	"os"
	"runtime"

	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

var DbRead, DbWrite *sql.DB
var DbReadSqlc, DbWriteSqlc *sqlc_model.Queries

//go:embed schema.sql
var ddl string

func Start() {
	os.MkdirAll("data", 0755)

	DbRead, _ = sql.Open("sqlite3", "file:data/hq.db")
	DbRead.SetMaxOpenConns(runtime.NumCPU())

	DbWrite, _ = sql.Open("sqlite3", "file:data/hq.db")
	DbWrite.SetMaxOpenConns(1)
	DbWrite.Exec("PRAGMA journal_mode=WAL")

	if _, err := DbWrite.Exec(ddl); err != nil {
		slog.Error("Failed to create schema", "error", err.Error()) // ignore this
	}

	DbReadSqlc = sqlc_model.New(DbRead)
	DbWriteSqlc = sqlc_model.New(DbWrite)
}

func Shutdown() {
	DbRead.Close()
	DbWrite.Close()
}
