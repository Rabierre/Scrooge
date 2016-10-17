package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rabierre/scrooge/models"
	"gopkg.in/gorp.v1"
)

var (
	db  *sql.DB
	dbm *gorp.DbMap
)

func main() {
	err := error(nil)
	db, err = sql.Open("sqlite3", "database")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	InitDB()

	r := NewEngine()
	r.Run()
}

func NewEngine() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/day/:date", func(c *gin.Context) {
		t, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			panic(err)
		}

		records := &[]models.Record{}
		_, err = dbm.Select(records, "select * from Record where Time = ?", t)
		if err != nil {
			panic(err)
		}

		rsByKind := sortByKind(records)

		c.HTML(http.StatusOK, "day.tmpl", gin.H{
			"time":        t.String(),
			"dayRecord":   records,
			"byKind":      rsByKind,
			"totalAmount": totalAmount(records),
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

		dbm.Insert(&models.Record{0, date, amount, kind})

		c.HTML(http.StatusCreated, "insert.tmpl", gin.H{})
	})

	router.POST("/update/:recordId", func(c *gin.Context) {
		id := c.Param("recordId")

		record := &models.Record{}
		err := dbm.SelectOne(record, "select * from Record where Id = ?", id)

		t, err := time.Parse("2006-01-02", c.PostForm("date"))
		if err == nil {
			record.Time = t
		}
		if c.PostForm("amount") != "" {
			record.Amount = c.PostForm("amount")
		}
		if c.PostForm("kind") != "" {
			record.Kind = c.PostForm("kind")
		}

		cnt, err := dbm.Update(record)
		if err != nil {
			panic(err)
		}
		if cnt > 0 {
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/day/%v", t))
		}
	})

	return router
}

func totalAmount(records *[]models.Record) float64 {
	total := float64(0)
	if records == nil {
		return total
	}

	for _, r := range *records {
		m, _ := strconv.ParseFloat(r.Amount, 64)
		total += m
	}
	return total
}

func sortByKind(records *[]models.Record) *map[string][]models.Record {
	result := make(map[string][]models.Record)
	if records == nil {
		return &result
	}

	for _, r := range *records {
		result[r.Kind] = append(result[r.Kind], r)
	}

	return &result
}

func totalAmountByKind(records *map[string][]models.Record) *map[string]float64 {
	result := make(map[string]float64)
	if records == nil {
		return &result
	}

	for k, rs := range *records {
		result[k] = totalAmount(&rs)
	}
	return &result
}

func InitDB() {
	dbm = &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	setColumnSizes := func(t *gorp.TableMap, colSizes map[string]int) {
		for col, size := range colSizes {
			t.ColMap(col).MaxSize = size
		}
	}

	t := dbm.AddTable(models.Record{}).SetKeys(true, "Id")
	setColumnSizes(t, map[string]int{
		"Time":   50,
		"Amount": 50,
		"Kind":   50,
	})

	dbm.TraceOn("[gorp]", log.New(os.Stdout, "sql:", log.Lmicroseconds))
	err := dbm.CreateTablesIfNotExists()
	if err != nil {
		panic(fmt.Sprintf("Fail to create tables: %+v", err))
	}
}
