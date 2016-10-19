package models

import (
	"time"
)

type Record struct {
	Id      uint64
	Time    time.Time
	Amount  string
	LabelId uint64
}
