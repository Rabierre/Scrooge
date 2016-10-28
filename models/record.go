package models

import (
	"time"

	"github.com/rabierre/scrooge/share"
)

type Record struct {
	Id      uint64
	Time    time.Time
	Amount  string
	LabelId uint64
}

func (r *Record) LabelName() string {
	obj, err := share.Dbm.Get(Label{}, r.LabelId)
	if err != nil {
		panic(err)
	}
	return obj.(*Label).Name
}
