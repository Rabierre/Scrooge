package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGrepRecordsByDate(t *testing.T) {
	date1, _ := time.Parse("2006-01-02", "2016-10-07")
	date2, _ := time.Parse("2006-01-02", "2016-10-08")
	records := []Record{
		{date1, "1000.0", "Food"},
		{date2, "1000.0", "Food"},
	}
	result := grepRecordsByDate(&records, date1)
	assert.Equal(t, len(result), 1)
}

func TestTotalAmount(t *testing.T) {
	date1, _ := time.Parse("2006-01-02", "2016-10-07")
	date2, _ := time.Parse("2006-01-02", "2016-10-08")
	records := []Record{
		{date1, "1000.0", "Food"},
		{date2, "1000.0", "Food"},
	}

	result := totalAmount(&records)
	assert.Equal(t, result, 2000.0)
}

func TestSortByKind(t *testing.T) {
	date1, _ := time.Parse("2006-01-02", "2016-10-07")
	date2, _ := time.Parse("2006-01-02", "2016-10-08")
	records := []Record{
		{date1, "1000.0", "Food"},
		{date2, "1000.0", "Study"},
	}

	result := sortByKind(&records)
	for _, vv := range *result {
		// assert.Equal(t, k, "")
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

func setup() {
	f, err := os.OpenFile("records.json", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file = f
}

func setdown() {
	file.Close()
}
