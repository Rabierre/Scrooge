package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rabierre/scrooge/models"
	"github.com/stretchr/testify/assert"
)

func setup() {
	err := error(nil)
	db, err = sql.Open("sqlite3", "testdb")
	if err != nil {
		panic(err)
	}
	InitDB()
}

func setdown() {
	queries := []string{}
	cur, _ := db.Query("select name from sqlite_master where type = 'table';")
	for cur.Next() {
		name := ""
		cur.Scan(&name)
		if name != "sqlite_sequence" {
			queries = append(queries, fmt.Sprintf("delete from %s;", name))
		}
	}
	queries = append(queries, "VACUUM")

	for _, q := range queries {
		db.Exec(q)
	}
	db.Close()
}

func dummyRecords() *[]models.Record {
	date1, _ := time.Parse("2006-01-02", "2016-10-07")
	date2, _ := time.Parse("2006-01-02", "2016-10-08")
	return &[]models.Record{
		{1, date1, "1000.0", "Food"},
		{2, date2, "1000.0", "Study"},
	}
}

func TestTotalAmount(t *testing.T) {
	records := dummyRecords()
	result := totalAmount(records)
	assert.Equal(t, result, 2000.0)
}

func TestSortByKind(t *testing.T) {
	result := sortByKind(&[]models.Record{})
	for _, vv := range *result {
		assert.Equal(t, len(vv), 0)
	}

	records := dummyRecords()
	result = sortByKind(records)
	for _, vv := range *result {
		assert.Equal(t, len(vv), 1)
	}
}

func TestRecordsByDay(t *testing.T) {
	setup()
	defer setdown()

	today, _ := time.Parse(time.RFC3339, "2016-10-31T00:00:00+09:00")
	tomorrow, _ := time.Parse(time.RFC3339, "2016-11-01T00:00:00+09:00")
	rs := []*models.Record{
		&models.Record{0, today, "1000", "Food"},
		&models.Record{0, tomorrow, "2000", "Food"},
	}
	for _, r := range rs {
		dbm.Insert(r)
	}

	records := recordsByDate(today)
	assert.Equal(t, len(*records), 1)
	assert.Equal(t, (*records)[0].Id, rs[0].Id)
}

func TestRecordsByMonth(t *testing.T) {
	setup()
	defer setdown()

	thisMonth, _ := time.Parse(time.RFC3339, "2016-10-31T23:59:59+09:00")
	nextMonth, _ := time.Parse(time.RFC3339, "2016-11-01T00:00:00+09:00")
	rs := []*models.Record{
		&models.Record{0, thisMonth, "1000", "Food"},
		&models.Record{0, nextMonth, "2000", "Food"},
	}
	for _, r := range rs {
		dbm.Insert(r)
	}

	records := recordsByMonth(thisMonth)
	assert.Equal(t, len(*records), 1)
	assert.Equal(t, (*records)[0].Id, rs[0].Id)
}

// This code came from gin-gonic/gin/routes_test.go
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func PerformPostRequest(r http.Handler, method, path string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAllRoutesExist(t *testing.T) {
	setup()
	defer setdown()

	routeTests := []struct {
		method           string
		location         string
		expectStatusCode uint32
	}{
		{"GET", "/day/2016-01-01", http.StatusNotFound},
		{"GET", "/insert", http.StatusNotFound},
		{"POST", "/insert", http.StatusNotFound},
	}

	r := NewEngine()

	for _, rt := range routeTests {
		w := PerformRequest(r, rt.method, rt.location)
		assert.NotEqual(t, rt.expectStatusCode, w.Code)
	}
}

func TestUpdateRecord(t *testing.T) {
	setup()
	defer setdown()

	r := NewEngine()

	UpdateRecord(t, r)
	PartialUpdateRecord(t, r)
}

func UpdateRecord(t *testing.T, r http.Handler) {
	// Prepare record
	record := &models.Record{0, time.Now(), "1000", "Food"}
	dbm.Insert(record)

	// Update record
	location := fmt.Sprintf("/update/%d", record.Id)
	body := strings.NewReader("date=2016-10-17&amount=2000&kind=Study")
	w := PerformPostRequest(r, "POST", location, body)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check record is updated
	dbm.SelectOne(record, "select * from Record where Id = ?", record.Id)
	assert.Equal(t, record.Amount, "2000")
	assert.Equal(t, record.Kind, "Study")
}

func PartialUpdateRecord(t *testing.T, r http.Handler) {
	// Prepare record
	record := &models.Record{0, time.Now(), "1000", "Food"}
	dbm.Insert(record)

	// Update record
	location := fmt.Sprintf("/update/%d", record.Id)
	body := strings.NewReader("kind=Study")
	w := PerformPostRequest(r, "POST", location, body)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check record is updated
	dbm.SelectOne(record, "select * from Record where Id = ?", record.Id)
	assert.Equal(t, record.Amount, "1000")
	assert.Equal(t, record.Kind, "Study")
}
