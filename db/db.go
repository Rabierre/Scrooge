package db

import (
	"database/sql"
	"gopkg.in/gorp.v1"
)

var (
	Db  *sql.DB
	Dbm *gorp.DbMap
)
