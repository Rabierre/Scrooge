package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGrepRecordsByDate(t *testing.T) {
	date1, _ := time.Parse("2006-01-02", "2016-10-07")
	date2, _ := time.Parse("2006-01-02", "2016-10-08")
	records := []record{
		{date1, "1000.0", "Food"},
		{date2, "1000.0", "Food"},
	}
	result := GrepRecordsByDate(&records, date1)
	assert.Equal(t, len(result), 1)
}
