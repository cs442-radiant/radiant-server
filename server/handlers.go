package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func GetRestaurant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantName := vars["restaurantName"]

	// Fix to defense SQL injection
	rows, err := database.Query(fmt.Sprintf("SELECT id FROM Restaurant WHERE name = \"%s\"", restaurantName))

	if err != nil {
		// Handle error
		fmt.Fprintln(w, err)
	} else {
		defer rows.Close()

		var id int
		var exists bool = false

		type response struct {
			Exists bool `json:"exists"`
		}

		for rows.Next() {
			err := rows.Scan(&id)

			if err != nil {
				log.Fatal(err)
				fmt.Fprintln(w, err)
			}

			if err := json.NewEncoder(w).Encode(response{Exists: true}); err != nil {
				log.Fatal(err)
			}
			exists = true

			break
		}

		if !exists {
			if err := json.NewEncoder(w).Encode(response{Exists: false}); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func GetBundle(w http.ResponseWriter, r *http.Request) {
	database.Query("SELECT * FROM ")

	fmt.Fprintln(w, "Get Bundle!")
}

func PostSample(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Post Sample!")
}
