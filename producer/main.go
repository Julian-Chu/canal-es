package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
)

func main() {
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "root"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dataSource := fmt.Sprintf("%s:%s@(%s:3306)/test", dbUser, dbPassword, dbHost)
	fmt.Println(dataSource)
	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Fatalf("failed to open database: %+v\n", err)
	}
	defer db.Close()

	for err := db.Ping(); err != nil; err = db.Ping() {
		time.Sleep(time.Second)
	}

	names := []string{"Tom", "Mary", "John"}
	mod := uint64(len(names))
	var cnt uint64
	for ; ; cnt++ {
		stat, err := db.Prepare("INSERT INTO login(name) VALUES (?)")
		if err != nil {
			log.Fatalf("failed to prepare stat: %+v\n", err)
		}
		username := names[cnt%mod]
		res, err := stat.Exec(username)
		if err != nil {
			log.Fatalf("failed to insert data: %+v\n", err)
		}
		insertId, _ := res.LastInsertId()
		log.Printf("insert user: %s, insert id: %v\n", username, insertId)
		time.Sleep(time.Second)
	}
}
