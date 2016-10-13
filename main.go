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

type Record struct {
	Time   time.Time
	Amount string
	Kind   string
}

var file *os.File

func main() {
	f, err := os.OpenFile("records.json", os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	file = f
	defer file.Close()

	r := NewEngine()
	r.Run()
}

func NewEngine() *gin.Engine {
	// sql.Open("sqlite3", "")

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/day/:date", func(c *gin.Context) {
		t, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(file)
		read := "" // TODO more grace
		for scanner.Scan() {
			read += scanner.Text()
		}

		var records []Record
		err = json.Unmarshal([]byte(read), &records)
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

	router.GET("/insert", func(c *gin.Context) {
		c.HTML(http.StatusOK, "insert.tmpl", gin.H{})
	})

	router.POST("/insert", func(c *gin.Context) {
		amount := c.PostForm("amount")
		kind := c.PostForm("kind")
		date, err := time.Parse("2006-01-02", c.PostForm("date"))
		if err != nil || kind == "" || amount == "" {
			c.HTML(http.StatusForbidden, "insert.tmpl", gin.H{
				"amount": amount,
				"kind":   kind,
				"date":   c.PostForm("date"),
			})
			return
		}

		record := Record{date, amount, kind}
		toJson, err := json.Marshal(record)
		if err != nil {
			panic(err)
		}

		fileInfo, err := file.Stat()
		if err != nil {
			panic(err)
		}
		toJson = append([]byte(",\n\t"), toJson...)
		toJson = append(toJson, []byte("\n]")...)
		file.WriteAt(toJson, fileInfo.Size()-2)

		c.HTML(http.StatusCreated, "insert.tmpl", gin.H{})
	})

	return router
}

func grepRecordsByDate(records *[]Record, date time.Time) []Record {
	result := []Record{}
	for _, r := range *records {
		if r.Time.Year() == date.Year() && r.Time.Month() == date.Month() &&
			r.Time.Day() == date.Day() {
			result = append(result, r)
		}
	}

	return result
}

func sortByKind(records *[]Record) *map[string][]Record {
	result := make(map[string][]Record)

	for _, r := range *records {
		result[r.Kind] = append(result[r.Kind], r)
	}

	return &result
}

func totalAmount(records *[]Record) float64 {
	total := float64(0)
	for _, r := range *records {
		m, _ := strconv.ParseFloat(r.Amount, 64)
		total += m
	}
	return total
}

func totalAmountByKind(records *map[string][]Record) map[string]float64 {
	result := make(map[string]float64)
	for k, rs := range *records {
		result[k] = totalAmount(&rs)
	}
	return result
}
