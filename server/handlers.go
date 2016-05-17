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

		type Response struct {
			Exists bool `json:"exists"`
		}

		for rows.Next() {
			err := rows.Scan(&id)

			if err != nil {
				log.Fatal(err)
				fmt.Fprintln(w, err)
			}

			if err := json.NewEncoder(w).Encode(Response{Exists: true}); err != nil {
				log.Fatal(err)
			}
			exists = true

			break
		}

		if !exists {
			if err := json.NewEncoder(w).Encode(Response{Exists: false}); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func GetBundle(w http.ResponseWriter, r *http.Request) {
	type Request struct {
		RestaurantName    string `json:"restaurantName"`
		BundleDescription string `json:"bundleDescription"`
	}

	var request Request

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		fmt.Fprintln(w, err)
		log.Fatal(err)
		return
	}

	rows, err := database.Query(fmt.Sprintf("SELECT id FROM Restaurant WHERE name = \"%s\"", request.RestaurantName))
	defer rows.Close()

	if err != nil {
		fmt.Fprintln(w, err)
		log.Fatal(err)
		return
	}

	var restaurantId int

	if rows.Next() {
		rows.Scan(&restaurantId)
	} else {
		log.Println("No such restaurant exists... creating new restaurant")
		rows, err := database.Query("SELECT MAX(id) FROM Restaurant")

		if err != nil {
			fmt.Fprintln(w, err)
			log.Fatal(err)
			return
		}

		var max int = 0

		if rows.Next() {
			rows.Scan(&max)
		}

		defer rows.Close()

		{
			var newRestaurantId = max + 1

			_, err := database.Query(
				fmt.Sprintf("INSERT INTO Restaurant VALUES(%d, \"%s\")", newRestaurantId, request.RestaurantName),
			)

			if err != nil {
				fmt.Fprintln(w, err)
				log.Fatal(err)
				return
			}

			restaurantId = newRestaurantId
		}
	}

	{
		rows, err := database.Query("SELECT MAX(id) FROM Bundle")

		if err != nil {
			fmt.Fprintln(w, err)
			log.Fatal(err)
			return
		}

		var max int = 0

		if rows.Next() {
			rows.Scan(&max)
		}

		{
			var newBundleId = max + 1

			_, err := database.Query(
				fmt.Sprintf("INSERT INTO Bundle VALUES(%d, %d, \"%s\")", newBundleId, restaurantId, request.BundleDescription),
			)

			if err != nil {
				fmt.Fprintln(w, err)
				log.Fatal(err)
				return
			}

			type Response struct {
				BundleId int `json:"bundleId"`
			}

			json.NewEncoder(w).Encode(Response{BundleId: newBundleId})
		}
	}
}

func PostSample(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Post Sample!")
}
