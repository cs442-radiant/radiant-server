package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"log"
	"net/http"
)

func httpHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Hey")
}

func main() {
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

		var id int
		rows, err := db.Query("select * from Record")
		if err != nil {
			log.Fatal(err)
		} else {
			for rows.Next() {
				err := rows.Scan(&id)
				if err != nil {
					log.Fatal(err)
				}
				log.Println(id)
			}
		}
	}

	defer db.Close()

	http.HandleFunc("/", httpHandler)
	http.ListenAndServe(":8100", nil)
}
