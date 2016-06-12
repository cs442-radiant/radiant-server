package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sjwhitworth/golearn/base"
	"log"
	"net/http"
	"strconv"
	"time"
)

func checkErr(err error, w http.ResponseWriter, msg string, code int) bool {
	if err != nil {
		errMsg := fmt.Sprintf("%s: %s", msg, err)
		log.Println(errMsg)

		if w != nil {
			http.Error(w, errMsg, code)
		}
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

		_, err = database.Exec(fmt.Sprintf("INSERT INTO Restaurant (id, name) VALUES(%d, \"%s\")", newRestaurantId, request.RestaurantName))
		if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
			return
		}

		restaurantId = newRestaurantId
	}

	maxIdRows, err := database.Query("SELECT MAX(id) FROM Bundle")
	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}

	defer maxIdRows.Close()

	var max int = 0

	if maxIdRows.Next() {
		maxIdRows.Scan(&max)
	}

	var newBundleId = max + 1

	_, err = database.Exec(fmt.Sprintf("INSERT INTO Bundle (id, restaurantId, description) VALUES(%d, %d, \"%s\")", newBundleId, restaurantId, request.BundleDescription))
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

type WiFiSample struct {
	SSID         string `json:"SSID"`
	BSSID        string `json:"BSSID"`
	Capabilities string `json:"capabilities"`
	Level        int    `json:"level"`
	Frequency    int    `json:"frequency"`
}

func PostSample(w http.ResponseWriter, r *http.Request) {
	log.Println("PostSample")

	type Request struct {
		BundleId  int          `json:"bundleId"`
		Timestamp time.Time    `json:"timestamp"`
		WiFiList  []WiFiSample `json:"WiFiList"`
	}

	checkAndReconnect()

	var request Request

	if err := json.NewDecoder(r.Body).Decode(&request); checkErr(err, w, "Bad request JSON format", http.StatusBadRequest) {
		return
	}

	wifiSampleStringBuf := new(bytes.Buffer)

	if err := json.NewEncoder(wifiSampleStringBuf).Encode(request.WiFiList); checkErr(err, w, "Bad request JSON format (WiFiList)", http.StatusBadRequest) {
		return
	}

	_, err := database.Exec("INSERT INTO Sample (bundleId, serverTimestamp, clientTimestamp, sample) VALUES(?, ?, ?, ?)", request.BundleId, time.Now(), request.Timestamp, wifiSampleStringBuf.String())
	if checkErr(err, w, "SQL query failed", http.StatusInternalServerError) {
		return
	}

	type Response struct {
	}

	var response Response

	if err := json.NewEncoder(w).Encode(response); checkErr(err, w, "Failed to encode response", http.StatusInternalServerError) {
		return
	}
}

func PostLearn(w http.ResponseWriter, r *http.Request) {
	log.Println("PostLearn")

	if isLearning {
		log.Println("Already in learning process ...")
		return
	}

	go learn()
}

func GetCurrentLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("GetCurrentLocation")

	checkAndReconnect()

	type Request struct {
		WiFiList []WiFiSample `json:"WiFiList"`
	}

	var request Request

	if err := json.NewDecoder(r.Body).Decode(&request); checkErr(err, w, "Bad request JSON format", http.StatusBadRequest) {
		return
	}

	if classifier != nil {
		log.Println("Classifier ready. Preparing for the prediction...")

		output := []string{}

		for _, BSSID := range BSSIDList {
			var level int = -100

			for _, AP := range request.WiFiList {
				if AP.BSSID == BSSID {
					level = AP.Level
				}
			}

			output = append(output, strconv.Itoa(level))
		}

		// Make a new instance
		specs := make([]base.AttributeSpec, len(BSSIDList)+1)

		allAttrs := testData.AllAttributes()

		instance := base.NewDenseInstances()

		for i, attr := range allAttrs {
			specs[i] = instance.AddAttribute(attr)

			if len(allAttrs)-1 == i {
				instance.AddClassAttribute(attr)
			}
		}

		instance.Extend(1)

		for i, _ := range BSSIDList {
			instance.Set(specs[i], 0, specs[i].GetAttribute().GetSysValFromString(output[i]))
		}

		log.Println("New instance: ")
		log.Println(instance)

		predictions := classifier.Predict(instance)
		log.Println("Predictions: ")
		log.Println(predictions)

		type Response struct {
			RestaurantName string `json:"restaurantName"`
		}

		var result string = "No proper restaurant"

		if predictions != nil {
			result = predictions.RowString(0)
		}

		if err := json.NewEncoder(w).Encode(Response{RestaurantName: result}); checkErr(err, w, "Failed to encode response", http.StatusInternalServerError) {
			return
		}
	} else {
		log.Println("Classifier not ready yet.")
		http.Error(w, "Classifier not ready yet.", http.StatusBadRequest)
	}
}
