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

	"github.com/rabierre/scrooge/db"
	"github.com/rabierre/scrooge/models"
	"github.com/stretchr/testify/assert"
)

func setup() {
	err := error(nil)
	db.Db, err = sql.Open("sqlite3", "testdb")
	if err != nil {
		panic(err)
	}
	InitDB()
}

func setdown() {
	queries := []string{}
	cur, _ := db.Db.Query("select name from sqlite_master where type = 'table';")
	for cur.Next() {
		name := ""
		cur.Scan(&name)
		if name != "sqlite_sequence" {
			queries = append(queries, fmt.Sprintf("delete from %s;", name))
		}
	}
	queries = append(queries, "VACUUM")

	for _, q := range queries {
		db.Db.Exec(q)
	}
	db.Db.Close()
}

func dummyRecords() *[]models.Record {
	date1, _ := time.Parse("2006-01-02", "2016-10-07")
	date2, _ := time.Parse("2006-01-02", "2016-10-08")
	label1 := models.Label{Id: 1, Name: "Food"}
	label2 := models.Label{Id: 2, Name: "Study"}
	return &[]models.Record{
		{1, date1, "1000.0", label1.Id},
		{2, date2, "1000.0", label2.Id},
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

func TestRecordsByLabelId(t *testing.T) {
	setup()
	defer setdown()

	today, _ := time.Parse(time.RFC3339, "2016-10-31T00:00:00+09:00")
	tomorrow, _ := time.Parse(time.RFC3339, "2016-11-01T00:00:00+09:00")
	label := &models.Label{Id: 0, Name: "Food"}
	db.Dbm.Insert(label)
	rs := []*models.Record{
		&models.Record{0, today, "1000", label.Id},
		&models.Record{0, tomorrow, "2000", label.Id},
	}
	for _, r := range rs {
		db.Dbm.Insert(r)
	}

	records := recordsByLabelId(label.Id, today)
	assert.Equal(t, len(*records), 1)
	assert.Equal(t, (*records)[0].Id, rs[0].Id)
}

func TestRecordsByDay(t *testing.T) {
	setup()
	defer setdown()

	today, _ := time.Parse(time.RFC3339, "2016-10-31T00:00:00+09:00")
	tomorrow, _ := time.Parse(time.RFC3339, "2016-11-01T00:00:00+09:00")
	label := &models.Label{Id: 0, Name: "Food"}
	db.Dbm.Insert(label)
	rs := []*models.Record{
		&models.Record{0, today, "1000", label.Id},
		&models.Record{0, tomorrow, "2000", label.Id},
	}
	for _, r := range rs {
		db.Dbm.Insert(r)
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
	label := models.Label{Id: 0, Name: "Food"}
	rs := []*models.Record{
		&models.Record{0, thisMonth, "1000", label.Id},
		&models.Record{0, nextMonth, "2000", label.Id},
	}
	for _, r := range rs {
		db.Dbm.Insert(r)
	}

	records := recordsByMonth(thisMonth)
	assert.Equal(t, len(*records), 1)
	assert.Equal(t, (*records)[0].Id, rs[0].Id)
}

func TestRecordsByYear(t *testing.T) {
	setup()
	defer setdown()

	thisYear, _ := time.Parse(time.RFC3339, "2016-12-31T23:59:59+09:00")
	nextYear, _ := time.Parse(time.RFC3339, "2017-01-01T00:00:00+09:00")
	label := &models.Label{Id: 0, Name: "Food"}
	db.Dbm.Insert(label)
	rs := []*models.Record{
		&models.Record{0, thisYear, "1000", label.Id},
		&models.Record{0, nextYear, "2000", label.Id},
	}
	for _, r := range rs {
		db.Dbm.Insert(r)
	}

	records := recordsByYear(thisYear)
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
		{"GET", "/", http.StatusNotFound},
		{"GET", "/day/2016-01-01", http.StatusNotFound},
		{"GET", "/month/2016-01", http.StatusNotFound},
		{"GET", "/year/2016", http.StatusNotFound},
		{"GET", "/label/1", http.StatusNotFound},
		{"GET", "/insert", http.StatusNotFound},
		{"POST", "/insert", http.StatusNotFound},
		{"POST", "/update/1", http.StatusNotFound},
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
	label1 := &models.Label{Id: 0, Name: "Food"}
	label2 := &models.Label{Id: 0, Name: "Study"}
	db.Dbm.Insert(label1)
	db.Dbm.Insert(label2)
	record := &models.Record{0, time.Now(), "1000", label1.Id}
	db.Dbm.Insert(record)

	// Update record
	location := fmt.Sprintf("/update/%d", record.Id)
	body := strings.NewReader(fmt.Sprintf("date=2016-10-17&amount=2000&labelId=%d", label2.Id))
	w := PerformPostRequest(r, "POST", location, body)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check record is updated
	db.Dbm.SelectOne(record, "select * from Record where Id = ?", record.Id)
	assert.Equal(t, record.Amount, "2000")
	assert.Equal(t, record.LabelId, label2.Id)
}

func PartialUpdateRecord(t *testing.T, r http.Handler) {
	// Prepare record
	label1 := &models.Label{Id: 0, Name: "Food"}
	label2 := &models.Label{Id: 0, Name: "Study"}
	db.Dbm.Insert(label1)
	db.Dbm.Insert(label2)
	record := &models.Record{0, time.Now(), "1000", label1.Id}
	db.Dbm.Insert(record)

	// Update record
	location := fmt.Sprintf("/update/%d", record.Id)
	body := strings.NewReader(fmt.Sprintf("labelId=%d", label2.Id))
	w := PerformPostRequest(r, "POST", location, body)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check record is updated
	db.Dbm.SelectOne(record, "select * from Record where Id = ?", record.Id)
	assert.Equal(t, record.Amount, "1000")
	assert.Equal(t, record.LabelId, label2.Id)
}
