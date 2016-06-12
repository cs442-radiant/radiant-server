package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

var isLearning bool = false

func learn() {
	learnInit()
	learnMain()
	learnTeardown()
}

func learnInit() {
	isLearning = true
}

func learnMain() {
	checkAndReconnect()
	rows, err := database.Query("SELECT sample FROM Sample")

	if checkErr(err, nil, "SQL query failed", -1) {
		return
	}

	defer rows.Close()

	var stringSample string

	type AP struct {
		SSID         string `json:"SSID"`
		BSSID        string `json:"BSSID"`
		Level        int    `json:"level"`
		Frequency    int    `json:"frequency"`
		Capabilities string `json:"capabilities"`
	}

	const fileName string = "local/result.csv"
	file, err := os.Create(fileName)
	defer file.Close()

	if err != nil {
		log.Println(fmt.Sprintf("Failed to create %s", fileName))
		return
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var sample []AP

	for rows.Next() {
		if checkErr(rows.Scan(&stringSample), nil, "Failed to scan", -1) {
			return
		}

		if checkErr(json.NewDecoder(bytes.NewBufferString(stringSample)).Decode(&sample), nil, "Failed to decode JSON", -1) {
			return
		}
	}

	log.Println("Learning complete")
}

func learnTeardown() {
	log.Println("Learn teardown")
	isLearning = false
}
