package main

import (
	"gopkg.in/redis.v5"
	"net/http"
	"encoding/json"
	"antigloss/logger"
	"os"
	"errors"
	"fmt"
	"bytes"
)

// Endpoint specifies postback information for what to postback to.
type Endpoint struct {
	Method string "json:method" // "GET", "POST"
	URL string "json:url" // http://sample_domain_endpoint.com/data?title={mascot}&image={location}&foo={bar}"
}

// Datum specifies a single point of data to postback with.
type Datum struct {
	Mascot string "json:mascot" // Gopher
	Location string "json:location" // https://blog.golang.org/gopher/gopher.png
}

// Request is a request from the database with data on how to postback.
type Request struct {
	Endpoint Endpoint "json:endpoint"
	Data []Datum "json:data"
}

// sendResponse sends postback data to a specified URL with the listed method.
func sendResponse(url string, method string, data []Datum) {
	var resp *http.Response
	var err error

	for _, datum := range data {
		if method != "GET" && method != "POST" {
			err = errors.New("Postback method " + method + " was not GET or POST.")
		} else {
			req, err := http.NewRequest(method, url,
				bytes.NewBuffer([]byte(`{"mascot": "` + datum.Mascot + `", "": "` + datum.Location + `}`)))
			if err == nil {
				req.Header.Set("Content-Type", "application/json")
				client := &http.Client{}
				resp, err = client.Do(req)
			}
		}

		if err != nil {
			logger.Error("Postback failed, error: ")
			logger.Error(fmt.Sprintf("%T", err))
		} else {
			logger.Info("Postback success, response: ")
			logger.Info(fmt.Sprintf("%T", resp))
		}
	}
}

// main runs forever. If errors occur and it exits, this should be restarted
// by our service.
func main() {
	// Redis setup
	client := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		Password: "SHI/hel7",
		DB: 0, // default DB
	})
	_, err := client.Ping().Result()
	if err != nil {
		logger.Error("Error connecting to Redis server: ")
		logger.Error(fmt.Sprintf("%T", err))
		os.Exit(1)
	}

	// Logger setup
	logger.Init("/var/log",
		400, // max log files
		20, // logfiles to delete
		100, // log file max size in MB
		false) // log Trace?

	// Main execution
	for {
		// Grab postback data from the database in blocking mode with a
		// timeout.
		postback := client.BLPop(10, "data")
		result, err := postback.Result()

		// If there's an error here, its a BLPop timeout from no new
		// data. Log and continue.
		if err != nil {
			logger.Info("Checked for new data.")
		} else {
			// Put the postback data into our defined structure.
			var request Request
			err := json.Unmarshal([]byte(result[1]), &request)

			// If the data is good, fork off a sendResponse and
			// look for more data.
			if err != nil {
				logger.Error("JSON data parsing failed, error: ")
				logger.Error(fmt.Sprintf("%T", err))
			} else {
				logger.Info("Queue size " + client.LLen("data").String())
				logger.Info("Sending postback to " + request.Endpoint.URL)
				go sendResponse(request.Endpoint.URL, request.Endpoint.Method, request.Data)
			}
		}
	}
}


