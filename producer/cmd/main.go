package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

func main() {
	db, err := sql.Open("mysql", "root:root@(127.0.0.1:3306)/test")
	if err != nil {
		log.Fatalf("failed to open database: %+v\n", err)
	}
	defer db.Close()

	for err := db.Ping(); err != nil; err = db.Ping() {
		time.Sleep(time.Second)
	}

	for {
		stat, err := db.Prepare("INSERT INTO login(name) VALUES (?)")
		if err != nil {
			log.Fatalf("failed to prepare stat: %+v\n", err)
		}
		username := "test user"
		res, err := stat.Exec(username)
		if err != nil {
			log.Fatalf("failed to insert data: %+v\n", err)
		}
		insertId, _ := res.LastInsertId()
		log.Printf("insert user: %s, insert id: %v\n", username, insertId)
		time.Sleep(time.Second)
	}
}
