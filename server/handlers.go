package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

func checkErr(err error, w http.ResponseWriter, msg string, code int) bool {
	if err != nil {
		errMsg := fmt.Sprintf("%s: %s", msg, err)
		log.Println(errMsg)
		http.Error(w, errMsg, code)
	}

	return err != nil
}

func GetRestaurant(w http.ResponseWriter, r *http.Request) {
	log.Println("GetRestaurant")

	checkAndReconnect()

	vars := mux.Vars(r)
	restaurantName := vars["restaurantName"]

	// Fix to defense SQL injection
	rows, err := database.Query(fmt.Sprintf("SELECT id FROM Restaurant WHERE name = \"%s\"", restaurantName))

	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}

	defer rows.Close()

	var id int
	var exists bool = false

	type Response struct {
		Exists bool `json:"exists"`
	}

	for rows.Next() {
		if checkErr(rows.Scan(&id),
			w, "Scan failed", http.StatusInternalServerError) {
			return
		}
		if checkErr(json.NewEncoder(w).Encode(Response{Exists: true}),
			w, "JSON encoding failed", http.StatusInternalServerError) {
			return
		}

		exists = true
		break
	}

	if !exists {
		if checkErr(json.NewEncoder(w).Encode(Response{Exists: false}),
			w, "JSON encoding failed", http.StatusInternalServerError) {
			return
		}
	}
}

func PostBundle(w http.ResponseWriter, r *http.Request) {
	log.Println("PostBundle")

	checkAndReconnect()

	type Request struct {
		RestaurantName    string `json:"restaurantName"`
		BundleDescription string `json:"bundleDescription"`
	}

	var request Request

	if checkErr(json.NewDecoder(r.Body).Decode(&request),
		w, "Bad request JSON format", http.StatusBadRequest) {
		return
	}

	log.Println(fmt.Sprintf("Received request: %+v", request))

	rows, err := database.Query(fmt.Sprintf("SELECT id FROM Restaurant WHERE name = \"%s\"", request.RestaurantName))
	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}

	defer rows.Close()

	var restaurantId int

	if rows.Next() {
		rows.Scan(&restaurantId)
	} else {
		log.Println("No such restaurant exists... creating new restaurant")

		rows, err := database.Query("SELECT MAX(id) FROM Restaurant")
		if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
			return
		}

		var max int = 0

		if rows.Next() {
			rows.Scan(&max)
		}

		defer rows.Close()

		var newRestaurantId = max + 1

		_, err = database.Query(fmt.Sprintf("INSERT INTO Restaurant VALUES(%d, \"%s\")", newRestaurantId, request.RestaurantName))
		if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
			return
		}

		restaurantId = newRestaurantId
	}

	rows, err = database.Query("SELECT MAX(id) FROM Bundle")
	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}

	var max int = 0

	if rows.Next() {
		rows.Scan(&max)
	}

	var newBundleId = max + 1

	_, err = database.Query(fmt.Sprintf("INSERT INTO Bundle VALUES(%d, %d, \"%s\")", newBundleId, restaurantId, request.BundleDescription))
	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}

	type Response struct {
		BundleId int `json:"bundleId"`
	}

	if checkErr(json.NewEncoder(w).Encode(Response{BundleId: newBundleId}),
		w, "JSON encoding failed", http.StatusInternalServerError) {
		return
	}
}

func PostSample(w http.ResponseWriter, r *http.Request) {
	log.Println("PostSample")

	checkAndReconnect()

	type Request struct {
		BundleId  int       `json:"bundleId"`
		Timestamp time.Time `json:"timestamp"`
		WiFiList  string    `json:"WiFiList"`
	}

	var request Request

	if err := json.NewDecoder(r.Body).Decode(&request); checkErr(err, w, "Bad request JSON format", http.StatusBadRequest) {
		return
	}

	// TODO: Add logic to validate WiFiList string (JSON)

	_, err := database.Query("INSERT INTO Sample VALUES(?, ?, ?, ?)", request.BundleId, time.Now(), request.Timestamp, request.WiFiList)
	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}
}
