package models

import (
	"time"

	"github.com/rabierre/scrooge/db"
)

type Record struct {
	Id      uint64
	Time    time.Time
	Amount  string
	LabelId uint64
}

func (r *Record) LabelName() string {
	obj, err := db.Dbm.Get(Label{}, r.LabelId)
	if err != nil {
		panic(err)
	}
	return obj.(*Label).Name
}
