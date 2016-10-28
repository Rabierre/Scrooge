package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rabierre/scrooge/models"
	"github.com/rabierre/scrooge/share"
	"gopkg.in/gorp.v1"
)

var (
	Dbm *gorp.DbMap
)

func main() {
	OpenDB()
	defer CloseDB()
	InitDB()

	Dbm = share.Dbm

	r := NewEngine()
	r.Run()
}

func NewEngine() *gin.Engine {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		t, err := time.Parse("2006-01-02", c.Query("date"))
		if err != nil {
			t = time.Now()
		}

		today := StartOfThisDay(t)
		tmrr := StartOfNextDay(t)
		dayAmount, err := Dbm.SelectFloat("select sum(Amount) from Record where Time >= ? AND Time < ?", today, tmrr)
		if err != nil {
			dayAmount = 0.0
		}

		thisMonth := StartOfThisMonth(t)
		nextMonth := StartOfNextMonth(t)
		monthAmount, err := Dbm.SelectFloat("select sum(Amount) from Record where Time >= ? AND Time < ?", thisMonth, nextMonth)
		if err != nil {
			monthAmount = 0.0
		}

		thisYear := StartOfThisYear(t)
		nextYear := StartOfNextYear(t)
		yearAmount, err := Dbm.SelectFloat("select sum(Amount) from Record where Time >= ? AND Time < ?", thisYear, nextYear)
		if err != nil {
			monthAmount = 0.0
		}

		prev := StartOfPrevDay(t)

		c.HTML(http.StatusOK, "main.tmpl", gin.H{
			"time":        t,
			"dayAmount":   dayAmount,
			"monthAmount": monthAmount,
			"yearAmount":  yearAmount,
			"yUrl":        fmt.Sprintf("/year/%d", today.Year()),
			"mUrl":        fmt.Sprintf("/month/%d-%02d", today.Year(), today.Month()),
			"dUrl":        fmt.Sprintf("/day/%d-%02d-%02d", today.Year(), today.Month(), today.Day()),
			"prevUrl":     fmt.Sprintf("/?date=%d-%02d-%02d", prev.Year(), prev.Month(), prev.Day()),
			"nextUrl":     fmt.Sprintf("/?date=%d-%02d-%02d", tmrr.Year(), tmrr.Month(), tmrr.Day()),
		})
	})

	router.GET("/day/:date", func(c *gin.Context) {
		t, err := time.Parse("2006-01-02", c.Param("date"))
		if err != nil {
			panic(err)
		}

		records := recordsByDate(t)
		rsByKind := sortByKind(records)
		prev := StartOfPrevDay(t)
		next := StartOfNextDay(t)

		c.HTML(http.StatusOK, "day.tmpl", gin.H{
			"time":        &t,
			"dayRecord":   records,
			"byKind":      rsByKind,
			"totalAmount": totalAmount(records),
			"kindAmount":  totalAmountByKind(rsByKind),
			"prevUrl":     fmt.Sprintf("/day/%d-%02d-%02d", prev.Year(), prev.Month(), prev.Day()),
			"nextUrl":     fmt.Sprintf("/day/%d-%02d-%02d", next.Year(), next.Month(), next.Day()),
		})
	})

	router.GET("/month/:date", func(c *gin.Context) {
		t, err := time.Parse("2006-01", c.Param("date"))
		if err != nil {
			panic(err)
		}

		records := recordsByMonth(t)
		rsByKind := sortByKind(records)
		prev := StartOfPrevMonth(t)
		next := StartOfNextMonth(t)

		c.HTML(http.StatusOK, "month.tmpl", gin.H{
			"time":        &t,
			"dayRecord":   records,
			"byKind":      rsByKind,
			"totalAmount": totalAmount(records),
			"kindAmount":  totalAmountByKind(rsByKind),
			"prevUrl":     fmt.Sprintf("/month/%d-%02d", prev.Year(), prev.Month()),
			"nextUrl":     fmt.Sprintf("/month/%d-%02d", next.Year(), next.Month()),
		})
	})

	router.GET("/year/:year", func(c *gin.Context) {
		t, err := time.Parse("2006", c.Param("year"))
		if err != nil {
			panic(err)
		}

		records := recordsByYear(t)
		rsByKind := sortByKind(records)
		prev := StartOfPrevYear(t)
		next := StartOfNextYear(t)

		c.HTML(http.StatusOK, "year.tmpl", gin.H{
			"time":        &t,
			"dayRecord":   records,
			"byKind":      rsByKind,
			"totalAmount": totalAmount(records),
			"kindAmount":  totalAmountByKind(rsByKind),
			"prevUrl":     fmt.Sprintf("/year/%d", prev.Year()),
			"nextUrl":     fmt.Sprintf("/year/%d", next.Year()),
		})
	})

	router.GET("/label/:id", func(c *gin.Context) {
		labelId, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			panic("No label id")
		}
		t, err := time.Parse("2006-01-02", c.Query("date"))
		if err != nil {
			t = time.Now()
		}

		records := recordsByLabelId(labelId, t)
		c.HTML(http.StatusOK, "label_day.tmpl", gin.H{
			"time":        &t,
			"dayRecord":   records,
			"totalAmount": totalAmount(records),
		})
	})

	router.GET("/insert", func(c *gin.Context) {
		labels := &[]models.Label{}
		Dbm.Select(labels, "select * from Label")

		c.HTML(http.StatusOK, "insert.tmpl", gin.H{
			"labels": labels,
		})
	})

	router.POST("/insert", func(c *gin.Context) {
		amount := c.PostForm("amount")
		labelId := c.PostForm("labelId")
		date, err := time.Parse("2006-01-02", c.PostForm("date"))
		if err != nil || labelId == "" || amount == "" {
			c.HTML(http.StatusForbidden, "insert.tmpl", gin.H{
				"amount":  amount,
				"labelId": labelId,
				"date":    c.PostForm("date"),
			})
			return
		}

		u, _ := strconv.ParseUint(labelId, 10, 64)
		Dbm.Insert(&models.Record{0, date, amount, u})

		c.HTML(http.StatusCreated, "insert.tmpl", gin.H{})
	})

	router.POST("/update/:recordId", func(c *gin.Context) {
		id := c.Param("recordId")

		obj, err := Dbm.Get(models.Record{}, id)
		record := obj.(*models.Record)

		t, err := time.Parse(time.RFC3339, c.PostForm("date"))
		if err == nil {
			record.Time = t
		}
		if c.PostForm("amount") != "" {
			record.Amount = c.PostForm("amount")
		}
		u, err := strconv.ParseUint(c.PostForm("labelId"), 10, 64)
		if err == nil {
			record.LabelId = u
		}

		cnt, err := Dbm.Update(record)
		if err != nil {
			panic(err)
		}
		if cnt > 0 {
			c.Redirect(http.StatusSeeOther, fmt.Sprintf("/day/%v", t))
		}
	})

	return router
}

func recordsByLabelId(labelId uint64, t time.Time) *[]models.Record {
	startOfToday := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	startOfTomorrow := time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
	records := &[]models.Record{}
	_, err := Dbm.Select(records, "select * from Record where LabelId = ? and Time >= ? and Time < ?", labelId, startOfToday, startOfTomorrow)
	if err != nil {
		panic(err)
	}
	return records
}

func recordsByDate(t time.Time) *[]models.Record {
	startOfToday := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	startOfTomorrow := StartOfNextDay(t)
	records := &[]models.Record{}
	_, err := Dbm.Select(records, "select * from Record where Time >= ? and Time < ?", startOfToday, startOfTomorrow)
	if err != nil {
		panic(err)
	}
	return records
}

func recordsByMonth(t time.Time) *[]models.Record {
	startOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	startOfNextMonth := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	records := &[]models.Record{}
	_, err := Dbm.Select(records, "select * from Record where Time >= ? and Time < ?", startOfMonth, startOfNextMonth)
	if err != nil {
		panic(err)
	}
	return records
}

func recordsByYear(t time.Time) *[]models.Record {
	startOfYear := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
	startOfNextYear := time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, t.Location())
	records := &[]models.Record{}
	_, err := Dbm.Select(records, "select * from Record where Time >= ? and Time < ?", startOfYear, startOfNextYear)
	if err != nil {
		panic(err)
	}
	return records
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

func sortByKind(records *[]models.Record) *map[uint64][]models.Record {
	result := make(map[uint64][]models.Record)
	if records == nil {
		return &result
	}

	for _, r := range *records {
		result[r.LabelId] = append(result[r.LabelId], r)
	}

	return &result
}

func totalAmountByKind(records *map[uint64][]models.Record) *map[uint64]float64 {
	result := make(map[uint64]float64)
	if records == nil {
		return &result
	}

	for k, rs := range *records {
		result[k] = totalAmount(&rs)
	}
	return &result
}
