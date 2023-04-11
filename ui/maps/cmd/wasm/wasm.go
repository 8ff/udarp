package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

var queryInProcess = false
var autoRefreshInterval = 0
var minutesSinceLastRefresh = 0

// parseQueryText parses the query text and returns a map of params
func parseQueryText(text string) map[string]interface{} {
	// Capitalize all text
	text = strings.ToUpper(text)
	params := make(map[string]interface{})
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		// Split line by ":"
		pair := strings.Split(line, ":")
		if len(pair) < 2 {
			continue
		}
		// Get key and value
		key := strings.TrimSpace(pair[0])
		values := strings.Split(pair[1], ",")
		for i, value := range values {
			// Remove all spaces even if there are multiple
			value = strings.ReplaceAll(value, " ", "")

			// If value is empty set it to ANY
			if value == "" {
				value = "ANY"
			}

			// If key is BANDS then strip all "m"
			if key == "BANDS" {
				value = strings.ReplaceAll(value, "M", "")
				// if value contains any value other than numeric then remove it
				if _, err := strconv.Atoi(value); err != nil {
					value = ""
				}
			}

			// if SENDER or RECEIVER contains string MYCALL then set it to ANY
			if key == "SENDER" || key == "RECEIVER" {
				if strings.Contains(value, "MYCALL") {
					value = "ANY"
				}
			}

			// If key is PAST_HOURS then check if value is valid
			if key == "PAST_HOURS" {
				if value == "ANY" {
					value = "3"
				}
				hours, err := strconv.Atoi(value)
				if err != nil {
					continue
				}
				if hours > 24 {
					value = "24"
				}
			}

			// Check for auto refresh
			if key == "REFRESH" {
				// If value is numeric, more than 1 and less than 60 then set autoRefreshInterval
				refreshInterval, err := strconv.Atoi(value)
				if err != nil {
					continue
				}

				if refreshInterval > 0 && refreshInterval < 60 {
					// fmt.Printf("Setting auto refresh interval to %d minutes\n", refreshInterval)
					autoRefreshInterval = refreshInterval
				}

			}

			values[i] = value
		}

		// Go over the values and remove any empty values
		nonEmptyValues := make([]string, 0, len(values))
		for _, value := range values {
			if value != "" {
				nonEmptyValues = append(nonEmptyValues, value)
			}
		}
		values = nonEmptyValues

		// Convert []values to string
		valuesString := strings.Join(values, ",")
		params[key] = valuesString

		// If any of the keys are missing then add them and set them to ANY
		if _, ok := params["BANDS"]; !ok {
			params["BANDS"] = "ANY"
		}
		if _, ok := params["SENDER"]; !ok {
			params["SENDER"] = "ANY"
		}
		if _, ok := params["RECEIVER"]; !ok {
			params["RECEIVER"] = "ANY"
		}
		if _, ok := params["LIMIT"]; !ok {
			params["LIMIT"] = "100"
		}
		if _, ok := params["MODE"]; !ok {
			params["MODE"] = "ANY"
		}
		if _, ok := params["PAST_HOURS"]; !ok {
			params["PAST_HOURS"] = "3"
		}
	}
	return params
}

/// NOT USED ?
// func handleQuerySubmit() {
// 	doc := js.Global().Get("document")

// 	// Call js function removeMap()
// 	removeMap := js.Global().Get("removeMap")
// 	removeMap.Invoke()

// 	// Call js function newMap()
// 	newMap := js.Global().Get("newMap")
// 	mapq := newMap.Invoke()

// 	// Run js function once() on mapq
// 	mapq.Call("once", "load", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		// fmt.Println("MAP_LOADED: RUNNING QUERY")

// 		// Load text by id from filterQuery
// 		filterQuery := doc.Call("getElementById", "filterQuery")
// 		text := filterQuery.Get("value").String()

// 		// Call parseQueryText() to get params
// 		params := parseQueryText(text)
// 		fmt.Println("PARAMS:", params)

// 		// Run loadByQuery and pass "api/geojson/reports" url and params

// 		loadByQuery := js.Global().Get("loadByQuery")
// 		loadByQuery.Invoke("api/geojson/reports", params)
// 		// loadByQuery.Invoke("api/geojson/reports", map[string]interface{}{
// 		// 	"bands": "10,20",
// 		// 	"limit": 100,
// 		// })
// 		return nil
// 	}))
// }

// setQueryButton listens for a click on the queryButton and reloads the map
//// NOT USED ?
// func setQueryButton() {
// 	doc := js.Global().Get("document")
// 	element := doc.Call("getElementById", "queryButton")
// 	element.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		handleQuerySubmit()
// 		return nil
// 	}))
// }

// Show error
func showError(msg string) {
	doc := js.Global().Get("document")
	queryResponse := doc.Call("getElementById", "queryResponse")
	queryResponse.Set("innerHTML", msg)
	queryResponse.Call("setAttribute", "class", js.ValueOf("queryResponse queryResponseError"))
}

// Show success
func showSuccess(msg string) {
	doc := js.Global().Get("document")
	queryResponse := doc.Call("getElementById", "queryResponse")
	queryResponse.Set("innerHTML", msg)
	queryResponse.Call("setAttribute", "class", js.ValueOf("queryResponse queryResponseSuccess"))
}

// Show info
func showInfo(msg string) {
	doc := js.Global().Get("document")
	queryResponse := doc.Call("getElementById", "queryResponse")
	queryResponse.Set("innerHTML", msg)
	queryResponse.Call("setAttribute", "class", js.ValueOf("queryResponse queryResponseInfo"))
}

// Function which calls js createLayer(layerName)
func createLayer(layerName, markerIconName string) {
	createLayer := js.Global().Get("createLayer")
	createLayer.Invoke(layerName, markerIconName)
}

// Function which calls js removeLayer(layerName)
func removeLayer(layerName string) {
	removeLayer := js.Global().Get("removeLayer")
	removeLayer.Invoke(layerName)
}

func pushQuery() {
	doc := js.Global().Get("document")
	queryInProcess = true

	showInfo("Querying server...")

	// Load text by id from filterQuery
	filterQuery := doc.Call("getElementById", "filterQuery")
	text := filterQuery.Get("value").String()

	// Call parseQueryText() to get params
	params := parseQueryText(text)
	// fmt.Println("PARAMS:", params)

	// Do a get request ourselves and pass data to loadByQuery
	// Convert params to map[string]string
	paramsString := make(map[string]string)
	for key, value := range params {
		paramsString[key] = value.(string)
	}
	// Do a get request
	go func() {
		// fmt.Println("PARAMSQ:", paramsString)
		res, err := getData("api/geojson/reports", paramsString)
		if err != nil {
			showError("Unable to query server: " + err.Error())
			queryInProcess = false
			return
		}
		// fmt.Println("RESPONSE:", res)

		// Parse res as json
		var data map[string]interface{}
		err = json.Unmarshal([]byte(res), &data)
		if err != nil {
			// TODO: check server response
			showError("Unable to parse server response: " + err.Error())
			queryInProcess = false
			return
		}
		// fmt.Println("DATA:", data)

		// Check if "features" is empty
		if len(data["features"].([]interface{})) == 0 {
			// Print error to .queryResponse
			showError("No results found")
			// TODO: Clear layers

			queryInProcess = false
			return
		}

		// Check if map is loaded by using map.Loaded()
		loaded := js.Global().Get("map").Call("loaded")
		if !loaded.Bool() {
			// Print error to .queryResponse
			showError("Map not loaded")
			queryInProcess = false
			return
		}

		// We need to go over all features which are in GeoJSON format and set corresponding modes to correct layer
		// ft8_layer, wspr_layer, udarp_layer, generic_layer
		// MODES: FT8/FT4, WSPR, UDARP, everything else is GENERIC
		// As we go over each feature, instead of parsing we can stringify it and use strings.Contains() to check if it contains "FT8" or "WSPR" or "UDARP"
		// If it contains "FT8" then we add it to ft8_layer
		// If it contains "WSPR" then we add it to wspr_layer
		// If it contains "UDARP" then we add it to udarp_layer
		// If it contains none of these then we add it to generic_layer

		layerFeatures := make(map[string][]interface{})
		for _, layerName := range []string{"ft8_layer", "ft4_layer", "wspr_layer", "udarp_layer", "generic_layer", "own_layer"} {
			layerFeatures[layerName] = []interface{}{}
		}

		// Modify this loop to collect features for each layer
		features := data["features"].([]interface{})
		for _, feature := range features {
			var layerName string
			// Convert feature.properties.mode to string
			featureString := fmt.Sprintf("%v", feature.(map[string]interface{})["properties"].(map[string]interface{})["mode"])

			// fmt.Printf("FEATURE: %v\n", featureString)

			switch featureString {
			case "FT8":
				layerName = "ft8_layer"
			case "FT4":
				layerName = "ft4_layer"
			case "WSPR":
				layerName = "wspr_layer"
			case "UDARP":
				layerName = "udarp_layer"
			case "OWN":
				layerName = "own_layer"
			default:
				layerName = "generic_layer"
			}

			layerFeatures[layerName] = append(layerFeatures[layerName], feature)
		}

		// Set the data for each layer after the loop
		for layerName, features := range layerFeatures {
			source := js.Global().Get("map").Call("getSource", layerName)
			featureCollection := map[string]interface{}{
				"type":     "FeatureCollection",
				"features": features,
			}
			jsonData, err := json.Marshal(featureCollection)
			if err != nil {
				showError("Unable to marshal feature collection: " + err.Error())
				queryInProcess = false
				return
			}
			source.Call("setData", js.Global().Get("JSON").Call("parse", string(jsonData)))
		}

		showSuccess("Query successful")
		queryInProcess = false
	}()
}

func runQuery() {
	doc := js.Global().Get("document")
	element := doc.Call("getElementById", "queryButton")
	element.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) interface{} {

		if queryInProcess {
			showError("Query in process")
			return nil
		}

		pushQuery()
		return nil
	}))

	// return nil
	// }))
}

// Function that preloads these layers "ft8_layer", "wspr_layer", "udarp_layer", "generic_layer"
func preloadLayers() {
	// Create a layer
	createLayer("ft8_layer", "ft8MarkerIcon")
	createLayer("ft4_layer", "ft4MarkerIcon")
	createLayer("wspr_layer", "wsprMarkerIcon")
	createLayer("udarp_layer", "udarpMarkerIcon")
	createLayer("generic_layer", "genericMarkerIcon")
	createLayer("own_layer", "ownMarkerIcon")
}

// Function that auto reruns the query
// It runs every 1 minute and checks how many minutes since last refresh, if equal or more than autoRefreshInterval then it runs the query
func refreshRunner() {
	// Create a ticker
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C

		if autoRefreshInterval == 0 {
			continue
		}

		if minutesSinceLastRefresh >= autoRefreshInterval {
			pushQuery()
			minutesSinceLastRefresh = 0
		}

		minutesSinceLastRefresh++
	}
}

func main() {
	// TODO: Load some test query as demo

	// Create a layer
	// createLayer("places", "ft8MarkerIcon")
	// createLayer("places", "wsprMarkerIcon")
	preloadLayers()

	go refreshRunner()

	// Listen for a click on queryButton and reload map
	// go setQueryButton()
	go runQuery()

	// Block forever
	select {}
}
