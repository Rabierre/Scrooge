package models

import "time"

type Record struct {
	Id     uint32
	Time   time.Time
	Amount string
	Kind   string
}
