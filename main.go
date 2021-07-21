package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	DBhost     = "db"
	DBuser     = "postgres-dev"
	DBpassword = "mysecretpassword"
	DBname     = "dev"
	Migration  = `CREATE TABLE IF NOT EXISTS bulletins (
		id serial PRIMARY KEY,
		author text NOT NULL,
		content text NOT NULL,
		created_at timestamp with time zone DEFAULT current_timestamp
	)`
)

type Bulletin struct {
	Author    string    `json:"author" binding:"required"`
	Content   string    `json:"content" binding:"required"`
	CreatedAt time.Time `json:"created_at" binding:"required"`
}

var db *sql.DB

func GetBulletins() ([]Bulletin, error) {
	const q = `SELECT author, content, created_at FROM bulletins ORDER BY created_at DESC LIMIT 100`
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}

	results := make([]Bulletin, 0)
	for rows.Next() {
		var author string
		var content string
		var createdAt time.Time

		err = rows.Scan(&author, &content, &createdAt)
		if err != nil {
			return nil, err
		}
		results = append(results, Bulletin{Author: author, Content: content, CreatedAt: createdAt})
	}
	return results, nil
}

func AddBulletin(bulletin Bulletin) error {
	const q = `INSERT INTO bulletins(author, content, created_at) VALUES ($1, $2, $3)`
	_, err := db.Exec(q, bulletin.Author, bulletin.Content, bulletin.CreatedAt)
	return err
}

func main() {
	var err error

	r := gin.Default()

	r.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "Hello! this is the first page. To See everything, go to /boards!")
	})

	r.GET("/board", func(context *gin.Context) {
		results, err := GetBulletins()
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"status": "Internal error:" + err.Error()})
			return
		}
		context.JSON(http.StatusOK, results)
	})

	r.POST("/board", func(context *gin.Context) {
		var b Bulletin
		b.CreatedAt = time.Now()
		// fmt.Println("The error is:",context.Bind(&b))
		if context.Bind(&b) == nil {
			fmt.Println("Context bound successfully with the bulletin object.")
			if err := AddBulletin(b); err != nil {
				context.JSON(http.StatusInternalServerError, gin.H{"status": "Internal error:" + err.Error()})
				return
			}
			context.JSON(http.StatusOK, gin.H{"status": "OK"})
		}
	})

	// dbInfo := fmt.Sprintf("host=%s port=%d  user=%s password=%s dbname=%s sslmode=disable", DBhost, 5432, DBuser, DBpassword, DBname)
	dbInfo := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", DBuser, DBpassword, DBhost, 5432, DBname)
	fmt.Println(dbInfo)
	db, err = sql.Open("postgres", dbInfo)

	if err != nil {
		fmt.Println("Whoops there was an error while opening DB. Check the database information again please.")
		panic(err)
	}

	defer db.Close()

	_, err = db.Query(Migration)
	if err != nil {
		fmt.Println("Running Migration...: ", Migration)
		log.Println("Failed to run migration: ", err.Error())
		return
	}

	log.Println("running on port 3001")
	if err := r.Run(":3001"); err != nil {
		panic(err)
	}
}
