package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

var database *sql.DB

func checkAndReconnect() {
	tryReconnect := false

	if database == nil {
		log.Println("Trying to connect to SQL DB...")
		tryReconnect = true
	} else if err := database.Ping(); err != nil {
		log.Println("Connection to DB is lost. Trying to reconnect...")
		tryReconnect = true
		database.Close()
	}

	if tryReconnect {
		for {
			db, err := connectDB()

			if err != nil {
				log.Println("Failed to connect to DB... retrying")
			} else {
				database = db
				return
			}
		}
	}
}

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "radiant:radiant@tcp(ec2-54-191-70-38.us-west-2.compute.amazonaws.com:3306)/radiant")

	if err != nil {
		log.Println(err)
	} else {
		log.Print("Starting SQL connection...")
	}

	err = db.Ping()
	if err != nil {
		log.Println(err)
	} else {
		log.Print("Connection succeed")
	}

	return db, err
}

func main() {
	checkAndReconnect()
	defer database.Close()

	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8100", router))
}
