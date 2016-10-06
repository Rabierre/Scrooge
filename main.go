package main

import (
	"bufio"
	// "database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	fmt.Printf("%+v \n", records)

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/day/*date", func(c *gin.Context) {
		t, err := time.Parse("2006-Jan-02", c.Param("date"))
		if err != nil {
			t = time.Now()
		}

		c.HTML(http.StatusOK, "day.tmpl", gin.H{
			"time":    t.String(),
			"records": records,
		})
	})

	router.Run()
}
