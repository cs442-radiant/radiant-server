package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
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

type Restaurant struct {
	Id   int
	Name string
}

func getBundleMap() (map[int]Restaurant, error) {
	rows, err := database.Query("SELECT bundle.id, restaurant.id, restaurant.name FROM Restaurant restaurant JOIN Bundle bundle ON bundle.restaurantId = restaurant.id")

	if err != nil {
		return nil, err
	}

	result := make(map[int]Restaurant)

	for rows.Next() {
		var bundleId int
		var restaurantId int
		var restaurantName string

		if err := rows.Scan(&bundleId, &restaurantId, &restaurantName); err != nil {
			return nil, err
		}

		result[bundleId] = Restaurant{Id: restaurantId, Name: restaurantName}
	}

	return result, nil
}

func learnMain() {
	checkAndReconnect()

	bundleMap, err := getBundleMap()
	if err != nil {
		log.Println("getBundleMap failed.")
		return
	}

	rowsForCount, err := database.Query("SELECT COUNT(*) FROM Sample LIMIT 10")

	if checkErr(err, nil, "SQL query failed", -1) {
		return
	}

	defer rowsForCount.Close()

	var numOfSamples int
	for rowsForCount.Next() {
		if checkErr(rowsForCount.Scan(&numOfSamples), nil, "Failed to scan", -1) {
			return
		}
	}

	rows, err := database.Query("SELECT bundleId, sample FROM Sample LIMIT 10")

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

	type OutputRow struct {
		APLevel  map[string]int
		bundleId int
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

	APMap := make(map[string]string)

	const numPartitions = 20
	samplesPerPartition := float64(numOfSamples) / float64(numPartitions)

	currentPartition := 0
	count := 0

	log.Println("Number of samples: ", numOfSamples)
	log.Println("Starting to extract the AP list...")

	outputSlice := []OutputRow{}

	for rows.Next() {
		var bundleId int
		var APList []AP
		APLevel := make(map[string]int)

		if checkErr(rows.Scan(&bundleId, &stringSample), nil, "Failed to scan", -1) {
			return
		}

		if checkErr(json.NewDecoder(bytes.NewBufferString(stringSample)).Decode(&APList), nil, "Failed to decode JSON", -1) {
			return
		}

		for i := range APList {
			APLevel[APList[i].BSSID] = APList[i].Level

			_, exists := APMap[APList[i].BSSID]
			if !exists {
				APMap[APList[i].BSSID] = APList[i].SSID
			}
		}

		outputSlice = append(outputSlice, OutputRow{APLevel: APLevel, bundleId: bundleId})

		count++

		currentNewPartition := math.Floor(float64(count) / samplesPerPartition)

		if currentNewPartition > float64(currentPartition) {
			currentPartition = int(currentNewPartition)

			log.Println(fmt.Sprintf("%d / %d", count, numOfSamples))
		}
	}

	log.Println("Finished extracting AP list.")
	log.Println("Number of APs: ", len(APMap))

	output := []string{}
	for BSSID, _ := range APMap {
		output = append(output, BSSID)
	}
	output = append(output, "restaurantId")
	writer.Write(output)

	log.Println("Starting to make CSV file...")
	for _, row := range outputSlice {
		output := []string{}

		for BSSID, _ := range APMap {
			level, exists := row.APLevel[BSSID]
			if exists {
				output = append(output, strconv.Itoa(level))
			} else {
				// Minimum value: -100
				output = append(output, "-100")
			}
		}

		output = append(output, strconv.Itoa(bundleMap[row.bundleId].Id))

		writer.Write(output)
	}

	log.Println("Finished making CSV file.")

	log.Println("Learning complete")
}

func learnTeardown() {
	log.Println("Learn teardown")
	isLearning = false
}
