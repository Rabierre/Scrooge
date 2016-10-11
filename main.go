package main

import (
	"bufio"
	// "database/sql"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	// _ "github.com/mattn/go-sqlite3"
	"github.com/gin-gonic/gin"
)

type record struct {
	Time   time.Time
	Amount string
	Kind   string
}

func main() {
	// sql.Open("sqlite3", "")
	f, err := os.Open("records.json")
	defer f.Close()

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(f)
	read := "" // TODO more grace

	for scanner.Scan() {
		read += scanner.Text()
	}

	var records []record
	err = json.Unmarshal([]byte(read), &records)
	if err != nil {
		panic(err)
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/day/:date", func(c *gin.Context) {
		t, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			panic(err)
		}

		rs := grepRecordsByDate(&records, t)
		rsByKind := sortByKind(&rs)

		c.HTML(http.StatusOK, "day.tmpl", gin.H{
			"time":        t.String(),
			"dayRecord":   rs,
			"byKind":      *rsByKind,
			"totalAmount": totalAmount(&rs),
			"kindAmount":  totalAmountByKind(rsByKind),
		})
	})

	router.Run()
}

func grepRecordsByDate(records *[]record, date time.Time) []record {
	result := []record{}
	for _, r := range *records {
		if r.Time.Year() == date.Year() && r.Time.Month() == date.Month() &&
			r.Time.Day() == date.Day() {
			result = append(result, r)
		}
	}

	return result
}

func sortByKind(records *[]record) *map[string][]record {
	result := make(map[string][]record)

	for _, r := range *records {
		result[r.Kind] = append(result[r.Kind], r)
	}

	return &result
}

func totalAmount(records *[]record) float64 {
	total := float64(0)
	for _, r := range *records {
		m, _ := strconv.ParseFloat(r.Amount, 64)
		total += m
	}
	return total
}

func totalAmountByKind(records *map[string][]record) map[string]float64 {
	result := make(map[string]float64)
	for k, rs := range *records {
		result[k] = totalAmount(&rs)
	}
	return result
}
