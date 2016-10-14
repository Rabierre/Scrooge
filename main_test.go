package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
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

// This code came from gin-gonic/gin/routes_test.go
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
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
