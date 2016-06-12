package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/sjwhitworth/golearn/base"
	"github.com/sjwhitworth/golearn/evaluation"
	"github.com/sjwhitworth/golearn/knn"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
)

var classifier *knn.KNNClassifier = nil

const k int = 5
const testSetProp float64 = 0.5
const csvFileName string = "local/result.csv"

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

func makeClassifier() (*knn.KNNClassifier, error) {
	rawData, err := base.ParseCSVToInstances(csvFileName, true)
	if err != nil {
		log.Println("Failed to open: ", csvFileName)
		return nil, err
	}

	cls := knn.NewKnnClassifier("euclidean", k)

	trainData, testData := base.InstancesTrainTestSplit(rawData, testSetProp)
	cls.Fit(trainData)

	predictions := cls.Predict(testData)

	confusionMat, err := evaluation.GetConfusionMatrix(testData, predictions)
	if err != nil {
		log.Println(fmt.Sprintf("Unable to get confusion matrix: %s", err.Error()))
		return nil, err
	}
	log.Println("\n", evaluation.GetSummary(confusionMat))

	return cls, nil
}

func learnMain() {
	checkAndReconnect()

	bundleMap, err := getBundleMap()
	if err != nil {
		log.Println("getBundleMap failed.")
		return
	}

	rowsForCount, err := database.Query("SELECT COUNT(*) FROM Sample")

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

	rows, err := database.Query("SELECT bundleId, sample FROM Sample")

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

	file, err := os.Create(csvFileName)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to create %s", csvFileName))
		return
	}

	writer := csv.NewWriter(file)

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
			// Leave the strongest 5 APs only
			/*if len(APLevel) < 5 {
				APLevel[APList[i].BSSID] = APList[i].Level
			} else {
				var smallestValue int = 0
				var smallestKey string = ""

				for k, v := range APLevel {
					if smallestValue > v {
						smallestKey = k
						smallestValue = v
					}
				}

				if smallestValue < APList[i].Level {
					delete(APLevel, smallestKey)
					APLevel[APList[i].BSSID] = APList[i].Level
				}
			}*/

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

	// For ordered iteration
	keys := []string{}
	for k, _ := range APMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	output := []string{}
	lineCount := 0
	for _, BSSID := range keys {
		output = append(output, BSSID)
		//output = append(output, strings.Replace(BSSID, ":", "", -1))
		//output = append(output, strconv.Itoa(lineCount))
		lineCount++
	}
	output = append(output, "restaurantId")
	writer.Write(output)

	log.Println("Starting to make CSV file...")
	for _, row := range outputSlice {
		output := []string{}

		for _, BSSID := range keys {
			level, exists := row.APLevel[BSSID]
			if exists {
				output = append(output, strconv.Itoa((level+100)*(level+100)))
			} else {
				// Minimum value: -100
				output = append(output, "0")
			}
		}

		output = append(output, bundleMap[row.bundleId].Name)

		writer.Write(output)
	}

	// Close file here
	writer.Flush()
	file.Close()

	log.Println("Finished making CSV file.")

	log.Println("Starting to make classifier...")

	classifier, err = makeClassifier()
	if err != nil {
		log.Println("Failed to create classifier.")
		return
	}

	log.Println("Classifier is built successfully.")

	log.Println("Learning complete")
}

func learnTeardown() {
	log.Println("Learn teardown")
	isLearning = false
}
