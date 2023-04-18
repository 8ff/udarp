package main

// TODO Cache reports to file every n minutes
// TODO: Calculate distance between sender and receiver in report
// TODO Show sender and receiver on map as different icon
// TODO, remove console debug
// TODO: Fix accepting CM or MM bands
// TODO Fix favicon
// TODO Fix mobile support
// TODO Finish auto refresh

import (
	"bytes"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/8ff/maidenhead"
	"github.com/8ff/udarp/pkg/misc"
	"github.com/dgraph-io/badger/v3"
	"github.com/minio/highwayhash"
)

// This contains a liste of sha hashes of each report, when we process it we need to check if its on the list, if so, skip it
var ReportHashHistory []string

type Config struct {
	DB     *badger.DB
	DBPath string
}

var GlobalConfig Config

type Report struct {
	Band            string   `json:"band"`
	Frequency       float64  `json:"frequency"`
	Mode            string   `json:"mode"`
	Receiver        string   `json:"receiver"`
	Sender          string   `json:"sender"`
	ReportTime      int64    `json:"report_time"`
	SNR             int      `json:"snr"`
	ReceiverLon     float64  `json:"receiver_lon"`
	ReceiverLat     float64  `json:"receiver_lat"`
	ReceiverLocator string   `json:"receiver_locator"`
	SenderLon       float64  `json:"sender_lon"`
	SenderLat       float64  `json:"sender_lat"`
	SenderLocator   string   `json:"sender_locator"`
	Power           int      `json:"power"`
	Drift           int      `json:"drift"`
	Tags            []string `json:"tags"`
	Distance        float64  `json:"distance"`
}

var CurrentReports []Report
var PSKReporterLog []ReceptionReports
var hashKey = make([]byte, 32)

// Define Geometry struct
type GeoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// Define Properties struct
type GeoJSONProperties struct {
	Description string `json:"description"`
	Mode        string `json:"mode"`
}

// Define Feature struct
type GeoJSONFeature struct {
	Type       string            `json:"type"`
	Geometry   GeoJSONGeometry   `json:"geometry"`
	Properties GeoJSONProperties `json:"properties"`
}

// Define GeoJSON struct
type GeoJSON struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONInput struct {
	Description string  `json:"description"`
	Mode        string  `json:"mode"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
}

type ReceptionReports struct {
	// XMLName        xml.Name `xml:"receptionReports"`
	// Text           string   `xml:",chardata"`
	CurrentSeconds string `xml:"currentSeconds,attr"`
	ActiveReceiver []struct {
		Text               string `xml:",chardata"`
		Callsign           string `xml:"callsign,attr"`
		Locator            string `xml:"locator,attr"`
		Frequency          string `xml:"frequency,attr"`
		Region             string `xml:"region,attr"`
		DXCC               string `xml:"DXCC,attr"`
		DecoderSoftware    string `xml:"decoderSoftware,attr"`
		AntennaInformation string `xml:"antennaInformation,attr"`
		Mode               string `xml:"mode,attr"`
		Bands              string `xml:"bands,attr"`
	} `xml:"activeReceiver"`
	LastSequenceNumber struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"lastSequenceNumber"`
	MaxFlowStartSeconds struct {
		Text  string `xml:",chardata"`
		Value string `xml:"value,attr"`
	} `xml:"maxFlowStartSeconds"`
	ReceptionReport []struct {
		Text               string  `xml:",chardata"`
		ReceiverCallsign   string  `xml:"receiverCallsign,attr"`
		ReceiverLocator    string  `xml:"receiverLocator,attr"`
		ReceiverLat        float64 `xml:"receiverLat,attr"`
		ReceiverLon        float64 `xml:"receiverLon,attr"`
		SenderCallsign     string  `xml:"senderCallsign,attr"`
		SenderLocator      string  `xml:"senderLocator,attr"`
		SenderLat          float64 `xml:"senderLat,attr"`
		SenderLon          float64 `xml:"senderLon,attr"`
		Frequency          string  `xml:"frequency,attr"`
		FlowStartSeconds   string  `xml:"flowStartSeconds,attr"`
		Mode               string  `xml:"mode,attr"`
		SenderDXCC         string  `xml:"senderDXCC,attr"`
		SenderDXCCCode     string  `xml:"senderDXCCCode,attr"`
		SenderDXCCLocator  string  `xml:"senderDXCCLocator,attr"`
		SenderLotwUpload   string  `xml:"senderLotwUpload,attr"`
		SenderEqslAuthGuar string  `xml:"senderEqslAuthGuar,attr"`
		SNR                string  `xml:"sNR,attr"`
	} `xml:"receptionReport"`
	ActiveCallsign []struct {
		Text      string `xml:",chardata"`
		Callsign  string `xml:"callsign,attr"`
		Reports   string `xml:"reports,attr"`
		DXCC      string `xml:"DXCC,attr"`
		DXCCcode  string `xml:"DXCCcode,attr"`
		Frequency string `xml:"frequency,attr"`
	} `xml:"activeCallsign"`
}

//go:embed www/*
var wwwContent embed.FS

func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const EarthRadius = 6371 // Earth's radius in km

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	lat1 = toRadians(lat1)
	lat2 = toRadians(lat2)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1)*math.Cos(lat2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return EarthRadius * c
}

func checkReportHash(report Report) bool {
	// If the ReportHashHistory len is more than 1M, clear it
	if len(ReportHashHistory) > 10000000 {
		ReportHashHistory = nil
	}

	// Generate the hash key if it's empty
	if len(hashKey) == 0 {
		hashKey = make([]byte, 32)
		_, err := rand.Read(hashKey)
		if err != nil {
			panic("Failed to generate a random key")
		}
	}

	// Create a HighwayHash instance
	h, err := highwayhash.New(hashKey)
	if err != nil {
		// Handle error
	}

	// Create a hash of the report
	reportString := fmt.Sprintf("%v", report)
	_, err = io.Copy(h, strings.NewReader(reportString))
	if err != nil {
		// Handle error
	}

	reportHash := h.Sum(nil)
	hashString := hex.EncodeToString(reportHash)

	// Check if the report hash is in the ReportHashHistory array
	for _, hash := range ReportHashHistory {
		if hash == hashString {
			// fmt.Printf("Found existing report hash: %x\n", reportHash)
			return true
		}
	}

	// Add the report hash to the ReportHashHistory array
	ReportHashHistory = append(ReportHashHistory, hashString)

	// If we get here, the report hash is not in the ReportHashHistory array
	return false
}

func parseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

func parseInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

// Server index.html on /
func serveIndex() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		index, err := wwwContent.ReadFile("www/index.html")
		if err != nil {
			panic(err)
		}
		w.Write(index)
	})
}

// Function which starts static file server and serves files with paths from wwwContent
func serveStatic() {
	fsys, err := fs.Sub(wwwContent, "www")
	if err != nil {
		panic(err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(fsys))))
}

// Function that calls url and returns xml
func getXml(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check exit code
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP status code %d", resp.StatusCode)
	}

	return body, nil
}

func packGeoJSONFeature(input GeoJSONInput) GeoJSONFeature {
	return GeoJSONFeature{
		Type: "Feature",
		Geometry: GeoJSONGeometry{
			Type: "Point",
			Coordinates: []float64{
				input.Lon,
				input.Lat,
			},
		},
		Properties: GeoJSONProperties{
			Description: input.Description,
			Mode:        input.Mode,
		},
	}
}

// Function that converts frequency to band in range between 135Khz to 76Ghz
func frequencyToBand(frequency string) string {
	// Convert frequency to float64
	f, err := strconv.ParseFloat(frequency, 64)
	if err != nil {
		return ""
	}

	// Convert frequency to band
	if f >= 135000 && f <= 137000 {
		return "2200m"
	}
	if f >= 472000 && f <= 479000 {
		return "630m"
	}
	if f >= 1800000 && f <= 2000000 {
		return "160m"
	}
	if f >= 3500000 && f <= 4000000 {
		return "80m"
	}
	if f >= 7000000 && f <= 7300000 {
		return "40m"
	}
	if f >= 10100000 && f <= 10150000 {
		return "30m"
	}
	if f >= 14000000 && f <= 14350000 {
		return "20m"
	}
	if f >= 18068000 && f <= 18168000 {
		return "17m"
	}
	if f >= 21000000 && f <= 21450000 {
		return "15m"
	}
	if f >= 24890000 && f <= 24990000 {
		return "12m"
	}
	if f >= 28000000 && f <= 29700000 {
		return "10m"
	}
	if f >= 50000000 && f <= 54000000 {
		return "6m"
	}
	if f >= 144000000 && f <= 148000000 {
		return "2m"
	}
	if f >= 430000000 && f <= 440000000 {
		return "70cm"
	}
	if f >= 1240000000 && f <= 1300000000 {
		return "23cm"
	}
	if f >= 2300000000 && f <= 2450000000 {
		return "13cm"
	}
	if f >= 3300000000 && f <= 3500000000 {
		return "9cm"
	}
	if f >= 5650000000 && f <= 5925000000 {
		return "6cm"
	}
	if f >= 10000000000 && f <= 10500000000 {
		return "3cm"
	}
	if f >= 24000000000 && f <= 24250000000 {
		return "1.25cm"
	}
	if f >= 47000000000 && f <= 47200000000 {
		return "6mm"
	}
	if f >= 75500000000 && f <= 81000000000 {
		return "4mm"
	}
	if f >= 119980000000 && f <= 120020000000 {
		return "2.5mm"
	}
	if f >= 142000000000 && f <= 149000000000 {
		return "2mm"
	}
	if f >= 241000000000 && f <= 250000000000 {
		return "1mm"
	}
	if f >= 470000000000 && f <= 500000000000 {
		return "0.6mm"
	}
	if f >= 750000000000 && f <= 1000000000000 {
		return "0.4mm"
	}
	if f >= 1200000000000 && f <= 1700000000000 {
		return "0.2mm"
	}
	if f >= 2300000000000 && f <= 3000000000000 {
		return "0.1mm"
	}
	if f >= 7600000000000 && f <= 10000000000000 {
		return "0.03mm"
	}
	if f >= 15000000000000 && f <= 30000000000000 {
		return "0.01mm"
	}
	if f >= 30000000000000 && f <= 50000000000000 {
		return "0.005mm"
	}
	if f >= 50000000000000 && f <= 100000000000000 {
		return "0.002mm"
	}
	if f >= 100000000000000 && f <= 300000000000000 {
		return "0.001mm"
	}
	if f >= 300000000000000 && f <= 500000000000000 {
		return "0.0005mm"
	}
	if f >= 500000000000000 && f <= 1000000000000000 {
		return "0.0002mm"
	}
	if f >= 1000000000000000 && f <= 3000000000000000 {
		return "0.0001mm"
	}
	if f >= 3000000000000000 && f <= 5000000000000000 {
		return "0.00005mm"
	}
	if f >= 5000000000000000 && f <= 10000000000000000 {
		return "0.00002mm"
	}
	if f >= 10000000000000000 && f <= 30000000000000000 {
		return "0.00001mm"
	}

	return ""
}

// Function that takes a string and string color and returns <span style=\"color: color;\">string</span>
func colorizeString(s string, color string) string {
	return fmt.Sprintf("<span style=\"color: %s;\">%s</span>", color, s)
}

func serveGeoJSONApi() {
	http.HandleFunc("/api/geojson/reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Read GET parameters
		// Get band
		bands := r.URL.Query().Get("BANDS")
		sender := r.URL.Query().Get("SENDER")
		receiver := r.URL.Query().Get("RECEIVER")
		mode := r.URL.Query().Get("MODE")
		past_hours := r.URL.Query().Get("PAST_HOURS")

		// Get limit parameter
		limit, err := strconv.Atoi(r.URL.Query().Get("LIMIT"))
		if err != nil {
			limit = 300
		}

		// TODO: Sanitize input
		fmt.Printf("Parameters: BANDS=%s, SENDER=%s, RECEIVER=%s, MODE=%s LIMIT=%d PAST_HOURS=%s\n", bands, sender, receiver, mode, limit, past_hours)

		// This is a json api
		w.Header().Set("Content-Type", "application/json")

		// Create geojson object
		geojson := GeoJSON{
			Type: "FeatureCollection",
		}

		// Create features array
		features := make([]GeoJSONFeature, 0)

		// Convert part hours to seconds
		past_hours_seconds, err := strconv.Atoi(past_hours)
		if err != nil {
			past_hours_seconds = 0
		}
		past_hours_seconds = past_hours_seconds * 3600

		reports, err := getReportsForLastNSeconds(past_hours_seconds)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Printf("Read %d reports from db\n", len(reports))

		gotSelfMarker := false
		// Iterate through all reports
		for _, report := range reports {
			// fmt.Printf("Band: %s Freq: %f Mode: %s Rec: %s Sen: %s ReportTime: %d, ReceiverLat: %f, ReceiverLon: %f, SenderLat: %f, SenderLon: %f\n", report.Band, report.Frequency, report.Mode, report.Receiver, report.Sender, report.ReportTime, report.ReceiverLat, report.ReceiverLon, report.SenderLat, report.SenderLon)

			// Check if report is in band
			if bands != "" && bands != "ANY" {
				bands := strings.Split(bands, ",")
				bandFound := false
				for _, b := range bands {
					if strings.EqualFold(report.Band, fmt.Sprintf("%sm", b)) {
						bandFound = true
						break
					}
				}
				if !bandFound {
					continue
				}
			}

			// Check if report is from sender
			if sender != "ANY" && sender != "" && !strings.EqualFold(report.Sender, sender) {
				continue
			}

			// Check if report is to receiver
			if receiver != "ANY" && receiver != "" && !strings.EqualFold(report.Receiver, receiver) {
				continue
			}

			// Check limit
			if len(features) >= limit {
				break
			}

			// Mode
			if mode != "ANY" && mode != "" {
				modes := strings.Split(mode, ",")
				modeFound := false
				for _, m := range modes {
					if strings.EqualFold(report.Mode, m) {
						modeFound = true
						break
					}
				}
				if !modeFound {
					fmt.Printf("Skipping mode: %s\n", report.Mode)
					continue
				}
			}

			// Check if report has valid coordinates
			if report.SenderLat == 0 || report.SenderLon == 0 {
				continue
			}

			// If no sender/receiver is specified, show all reports
			if (sender == "ANY" || sender == "") && (receiver == "ANY" || receiver == "") {
				feature := packGeoJSONFeature(GeoJSONInput{Mode: report.Mode, Description: fmt.Sprintf("<b>%s@%s<br>%s  %s  %s</b>", colorizeString(report.Receiver, "#ff5520"), colorizeString(report.ReceiverLocator, "#ffb500"), colorizeString(report.Band, "#4fff71"), colorizeString(report.Mode, "#00ffff"), colorizeString(fmt.Sprintf("%d", report.SNR), "#ff36d0")), Lat: report.ReceiverLat, Lon: report.ReceiverLon})
				features = append(features, feature)
			} else {
				// Satisfy OWN layer, find 1 report that has
				if !gotSelfMarker {
					// Find which sender or receiver is not empty or set to any
					// Set description for own marker
					if report.Sender != "" && report.Sender != "ANY" {
						feature := packGeoJSONFeature(GeoJSONInput{Mode: "OWN", Description: fmt.Sprintf("<b>%s</b>", colorizeString(report.Sender, "#3fffb5")), Lat: report.SenderLat, Lon: report.SenderLon})
						features = append(features, feature)
						gotSelfMarker = true
					} else if report.Receiver != "" && report.Receiver != "ANY" {
						feature := packGeoJSONFeature(GeoJSONInput{Mode: "OWN", Description: fmt.Sprintf("<b>%s</b>", colorizeString(report.Receiver, "#3fffb5")), Lat: report.ReceiverLat, Lon: report.ReceiverLon})
						features = append(features, feature)
						gotSelfMarker = true
					}
				}
				// Set description for report marker
				feature := packGeoJSONFeature(GeoJSONInput{Mode: report.Mode, Description: fmt.Sprintf("<b>%s@%s<br>%skm %s %s %s</b>", colorizeString(report.Receiver, "#ff5520"), colorizeString(report.ReceiverLocator, "#ffb500"), colorizeString(fmt.Sprintf("%.0f", report.Distance), "#ff7d00"), colorizeString(report.Band, "#4fff71"), colorizeString(report.Mode, "#00ffff"), colorizeString(fmt.Sprintf("%d", report.SNR), "#ff36d0")), Lat: report.ReceiverLat, Lon: report.ReceiverLon})
				features = append(features, feature)
			}
		}

		// Print all returned features
		// for _, feature := range features {
		// 	fmt.Printf("Feature: %s %f %f %s %s\n", feature.Type, feature.Geometry.Coordinates[0], feature.Geometry.Coordinates[1], feature.Properties.Mode, feature.Properties.Description)
		// }

		// Print how many features we have
		fmt.Printf("Returning results: %d\n", len(features))

		// Set features array to geojson object
		geojson.Features = features

		// Serve geojson as json
		json.NewEncoder(w).Encode(geojson)

	})
}

func ProcessPSKReporterReports111(reportsData ReceptionReports) []Report {
	var result []Report

	for _, report := range reportsData.ReceptionReport {
		freqF64, err := strconv.ParseFloat(report.Frequency, 64)
		if err != nil {
			fmt.Printf("Input: %s\n", report.Frequency)
			fmt.Println("Error parsing frequency:", err)
			continue
		}

		reportTime, err := strconv.ParseInt(report.FlowStartSeconds, 10, 64)
		if err != nil {
			fmt.Println("Error parsing report_time:", err)
			continue
		}
		snr, err := strconv.Atoi(report.SNR)
		if err != nil {
			fmt.Println("Error parsing SNR:", err)
			continue
		}

		r := Report{
			Band:        frequencyToBand(report.Frequency),
			Frequency:   freqF64,
			Mode:        report.Mode,
			Receiver:    report.ReceiverCallsign,
			Sender:      report.SenderCallsign,
			ReportTime:  reportTime,
			SNR:         snr,
			ReceiverLon: report.ReceiverLon,
			ReceiverLat: report.ReceiverLat,
			SenderLon:   report.SenderLon,
			SenderLat:   report.SenderLat,
			Power:       0,
			Drift:       0,
			Tags:        nil,
		}

		result = append(result, r)
	}

	return result
}

func ProcessWSPRReports(reportChan chan []Report, query string) error {
	processedReports := make([]Report, 0)

	resp, e := http.Get("http://db1.wspr.live/?query=" + url.QueryEscape(query))
	if e != nil {
		return e
	}
	defer resp.Body.Close()
	body, e := io.ReadAll(resp.Body)
	if e != nil {
		return e
	}

	lines := strings.Split(string(body), "\n")
	misc.Log("debug", fmt.Sprintf("Fetched %d WSPR reports", len(lines)-1))
	// First line contains these fields at these indexes:
	for i, line := range lines {

		// Print i every 1000 lines
		if i%1000 == 0 {
			misc.Log("debug", fmt.Sprintf("WSPR Processing line %d", i))
		}

		tokens := strings.Split(line, ",")
		if line == "" { // Skip empty lines
			continue
		}

		// If format changed, exit
		if len(tokens) < 20 || len(tokens) > 20 {
			fmt.Printf("returned CSV contains or more than 20 tokens, check returned CSV format: %s\n", tokens)
			continue
		}
		var e error
		wsprEntry := Report{}

		layout := "\"2006-01-02 15:04:05\""
		utcTime, e := time.Parse(layout, tokens[1])
		if e != nil {
			fmt.Printf("unable to parse time format: %s\n", e)
			continue
		}

		wsprEntry.ReportTime = utcTime.Unix()

		wsprEntry.Receiver = strings.Trim(tokens[3], "\"")
		wsprEntry.ReceiverLat, e = strconv.ParseFloat(tokens[4], 64)
		if e != nil {
			fmt.Printf("unable to parse rx_lat: %s", tokens[19])
			continue
		}

		wsprEntry.ReceiverLon, e = strconv.ParseFloat(tokens[5], 64)
		if e != nil {
			fmt.Printf("unable to parse rx_lon: %s\n", tokens[19])
			continue
		}

		wsprEntry.ReceiverLocator = strings.Trim(tokens[6], "\"")

		wsprEntry.Sender = strings.Trim(tokens[7], "\"")

		wsprEntry.SenderLat, e = strconv.ParseFloat(tokens[8], 64)
		if e != nil {
			fmt.Printf("unable to parse tx_lat: %s\n", tokens[19])
			continue
		}

		wsprEntry.SenderLon, e = strconv.ParseFloat(tokens[9], 64)
		if e != nil {
			fmt.Printf("unable to parse tx_lon: %s\n", tokens[19])
			continue
		}

		wsprEntry.SenderLocator = strings.Trim(tokens[10], "\"")

		wsprEntry.Frequency, e = strconv.ParseFloat(tokens[14], 64)
		if e != nil {
			fmt.Printf("unable to parse frequency: %s\n", tokens[19])
			continue
		}

		wsprEntry.Power, e = strconv.Atoi(tokens[15])
		if e != nil {
			fmt.Printf("unable to parse power: %s\n", tokens[19])
			continue
		}

		wsprEntry.SNR, e = strconv.Atoi(tokens[16])
		if e != nil {
			fmt.Printf("unable to parse snr: %s\n", tokens[19])
			continue
		}

		wsprEntry.Drift, e = strconv.Atoi(tokens[17])
		if e != nil {
			fmt.Printf("unable to parse drift: %s\n", tokens[19])
			continue
		}

		wsprEntry.Band = frequencyToBand(fmt.Sprintf("%f", wsprEntry.Frequency))
		if wsprEntry.Band == "" {
			wsprEntry.Band = tokens[2]
		}

		wsprEntry.Mode = "WSPR" // Assuming the mode is always WSPR

		wsprEntry.Distance = calculateDistance(wsprEntry.SenderLat, wsprEntry.SenderLon, wsprEntry.ReceiverLat, wsprEntry.ReceiverLon)

		// Send to channel
		if !checkReportHash(wsprEntry) {
			processedReports = append(processedReports, wsprEntry)
		}
	}

	misc.Log("debug", "WSPR Report processing done.")
	reportChan <- processedReports
	misc.Log("debug", "WSPR Reports sent to channel.")

	return nil
}

func ProcessPSKReporterReports(reportChan chan []Report) error {
	processedReports := make([]Report, 0)
	var reports ReceptionReports

	xmlData, err := getXml("https://retrieve.pskreporter.info/query")
	if err != nil {
		return err
	}

	err = xml.Unmarshal(xmlData, &reports)
	if err != nil {
		return err
	}

	misc.Log("debug", fmt.Sprintf("Fetched %d PSKReporter reports", len(reports.ReceptionReport)))
	for _, report := range reports.ReceptionReport {

		if len(report.ReceiverLocator) > 6 {
			report.ReceiverLocator = report.ReceiverLocator[:6]
		}
		// If empty then skip
		if report.ReceiverLocator == "" {
			continue
		}

		receiverLat, receiverLon, err := maidenhead.GetCoordinates(report.ReceiverLocator)
		if err != nil {
			fmt.Printf("RECEIVER > REC_LOC: %s, SENDER_LOC: %s\n", report.ReceiverLocator, report.SenderLocator)
			continue
		}

		// Lookup SenderLat and SenderLon
		// If report.SenderLocator is more than 6 chars them trim it to 6 chars
		if len(report.SenderLocator) > 6 {
			report.SenderLocator = report.SenderLocator[:6]
		}

		// If empty then skip
		if report.SenderLocator == "" {
			continue
		}

		senderLat, senderLon, err := maidenhead.GetCoordinates(report.SenderLocator)
		if err != nil {
			fmt.Printf("SENDER > REC_LOC: %s, SENDER_LOC: %s\n", report.ReceiverLocator, report.SenderLocator)
			continue
		}

		report := Report{
			Band:            frequencyToBand(report.Frequency),
			Frequency:       parseFloat64(report.Frequency),
			Mode:            report.Mode,
			Receiver:        report.ReceiverCallsign,
			Sender:          report.SenderCallsign,
			ReportTime:      parseInt64(report.FlowStartSeconds),
			SNR:             parseInt(report.SNR),
			ReceiverLon:     receiverLon,
			ReceiverLat:     receiverLat,
			ReceiverLocator: report.ReceiverLocator,
			SenderLon:       senderLon,
			SenderLat:       senderLat,
			SenderLocator:   report.SenderLocator,
			Power:           0, // Power is not provided in the ReceptionReports struct
			Drift:           0, // Drift is not provided in the ReceptionReports struct
			Tags:            []string{},
			Distance:        calculateDistance(senderLat, senderLon, receiverLat, receiverLon),
		}

		if !checkReportHash(report) {
			// Send to channel
			processedReports = append(processedReports, report)
		}
	}

	reportChan <- processedReports
	misc.Log("debug", "PSKReporter Reports processed")

	return nil
}

func getReportsByTime(db *badger.DB, startTime, endTime int64) ([]Report, error) {
	var reports []Report

	err := db.View(func(txn *badger.Txn) error {
		startKey := fmt.Sprintf("reports_%d", startTime)
		endKey := fmt.Sprintf("reports_%d", endTime)

		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.Prefix = nil

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(startKey)); it.Valid(); it.Next() {
			item := it.Item()
			key := item.Key()

			// Check if the key is less than the end key
			if bytes.Compare(key, []byte(endKey)) >= 0 {
				break
			}

			value, err := item.ValueCopy(nil)
			if err != nil {
				return fmt.Errorf("failed to copy value from BadgerDB item: %w", err)
			}

			var storedReports []Report
			err = json.Unmarshal(value, &storedReports)
			if err != nil {
				return fmt.Errorf("failed to unmarshal data from BadgerDB: %w", err)
			}

			reports = append(reports, storedReports...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return reports, nil
}

func storeReports(reportChan chan []Report, db *badger.DB) {
	var reports []Report
	ticker := time.NewTicker(1 * time.Minute)

	for {
		select {
		case reportBatch := <-reportChan:
			reports = append(reports, reportBatch...)
		case <-ticker.C:
			if len(reports) > 0 {
				// Store reports to the database
				err := db.Update(func(txn *badger.Txn) error {
					tsKey := fmt.Sprintf("reports_%d", time.Now().Unix())
					data, err := json.Marshal(reports)
					if err != nil {
						return err
					}

					entry := badger.NewEntry([]byte(tsKey), data).WithTTL(48 * time.Hour)
					return txn.SetEntry(entry)
				})

				if err != nil {
					log.Printf("Error storing reports: %v", err)
				} else {
					log.Printf("Stored %d reports to the database", len(reports))
					reports = nil // Clear reports slice
				}
			}
		}
	}
}

func getReportsForLastNSeconds(nSeconds int) ([]Report, error) {
	// Get the current time and the time n seconds ago
	now := time.Now()
	nSecondsAgo := now.Add(time.Duration(-nSeconds) * time.Second)

	// Convert them to Unix timestamps
	startTime := nSecondsAgo.Unix()
	endTime := now.Unix()

	// Get records for the last n seconds
	reports, err := getReportsByTime(GlobalConfig.DB, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("error getting reports: %w", err)
	}

	return reports, nil
}

func countRecords(db *badger.DB) (int, error) {
	var count int

	err := db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			count++
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return count, nil
}

func main() {
	GlobalConfig = Config{
		DBPath: "/tmp/udarpdb",
	}

	var err error
	opts := badger.DefaultOptions(GlobalConfig.DBPath)
	opts.Logger = nil
	GlobalConfig.DB, err = badger.Open(opts)
	if err != nil {
		log.Fatalf("Failed to open BadgerDB: %v", err)
	}
	defer GlobalConfig.DB.Close()

	// Create a channel to store reports
	ReportChan := make(chan []Report)
	go storeReports(ReportChan, GlobalConfig.DB)

	// Get reports every minute
	go func() {
		for {
			err := ProcessPSKReporterReports(ReportChan)
			if err != nil {
				log.Printf("Error getting PSKReporter reports: " + err.Error())
			}
			time.Sleep(2 * time.Minute)
		}
	}()

	// Get WSPR reports every minute
	go func() {
		fetchTime := 60 * time.Second
		queryTime := 300 // Seconds
		for {
			err := ProcessWSPRReports(ReportChan, fmt.Sprintf("SELECT * FROM wspr.rx where time > subtractSeconds(now(), %d) order by time desc FORMAT CSV;", queryTime))
			if err != nil {
				log.Printf("Error getting WSPR reports: " + err.Error())
			}
			time.Sleep(fetchTime)
		}
	}()

	serveStatic() // serveStatic()
	serveIndex()
	serveGeoJSONApi()
	http.ListenAndServe(":80", nil)
}
