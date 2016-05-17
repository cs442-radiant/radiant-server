package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
)

var database *sql.DB

func connectDB() (*sql.DB, error) {
	db, err := sql.Open("mysql", "radiant:radiant@tcp(ec2-54-191-70-38.us-west-2.compute.amazonaws.com:3306)/radiant")

	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("SQL connection")
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Connection success")
	}

	return db, err
}

func main() {
	var err error
	database, err = connectDB()

	if err != nil {
		panic("Failed to connect to DB")
	}

	defer database.Close()

	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8100", router))
}
