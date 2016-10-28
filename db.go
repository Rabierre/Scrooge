package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/rabierre/scrooge/models"
	"github.com/rabierre/scrooge/share"
	"gopkg.in/gorp.v1"
)

func OpenDB() {
	// err := error(nil)
	_db, err := sql.Open("sqlite3", "database")
	if err != nil {
		log.Fatal(err)
	}
	share.Db = _db
}

func CloseDB() {
	share.Db.Close()
}

func InitDB() {
	share.Dbm = &gorp.DbMap{Db: share.Db, Dialect: gorp.SqliteDialect{}}
	share.Dbm.TraceOn("[gorp]", log.New(os.Stderr, "sql:", log.Lmicroseconds))

	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}

	t := share.Dbm.AddTable(models.Record{}).SetKeys(true, "Id")
	setColumnSizes(t, map[string]int{
		"Time":    50,
		"Amount":  50,
		"LabelId": 50,
	})
	t = share.Dbm.AddTable(models.Label{}).SetKeys(true, "Id")
	setColumnSizes(t, map[string]int{
		"Name":       50,
		"CategoryId": 50,
	})
	t = share.Dbm.AddTable(models.Category{}).SetKeys(true, "Id")
	setColumnSizes(t, map[string]int{
		"Name": 50,
	})

	err := share.Dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(fmt.Sprintf("Fail to create tables: %+v", err))
	}
}
