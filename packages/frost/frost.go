package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

// how to represent location
type Request struct {
	Lat string `json:"lat"`
	Lon string `json:"lon"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}
type ForecastRequest struct {
	Cod     string
	Message float64
	Cnt     int
	List    []Measurement
}

type Measurement struct {
	Dt   int
	Main Weather
}

type Weather struct {
	Dt_txt  string
	TempMin float64 `json:"temp_min,omitempty"`
}

// Help from https://www.omnicalculator.com/other/mayan-calendar#how-to-convert-a-date-to-the-long-count-calendar

// Calculators tools
// https://utahgeology.com/bin/maya-calendar-converter/
// https://maya.nmai.si.edu/calendar/maya-calendar-converter

func Main(req Request) (*Response, error) {
	//hash of month and day lengths for validation
	//	monthLengths := map[int]int{0: 31, 1: 29, 2: 31, 3: 30, 4: 31, 5: 30, 6: 31, 7: 31, 8: 30, 9: 31, 10: 30, 11: 31}
	//	months := map[int]string{0: "January", 1: "February", 2: "March", 3: "April", 4: "May", 5: "June", 6: "July", 7: "August", 8: "September", 9: "October", 10: "November", 11: "December"}

	// Used for local testing
	if req.Lat == "" || req.Lon == "" {
		return nil, errors.New("Lat and Lon parameters are required")
	}

	lat, errla := strconv.ParseFloat(req.Lat, 64)
	lon, errlon := strconv.ParseFloat(req.Lon, 64)

	// Set API token in serverless Env
	apiKey := os.Getenv("APIKEY")

	if errla != nil || errlon != nil {
		return nil, fmt.Errorf("Error parsing lat and lon as floats,", errla, errlon)
	}

	query := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast/hourly?lat=%f&lon=%f&appid=%s&mode=json&cnt=96", lat, lon, apiKey)
	r, err := http.Get(query)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error returned from %s: %s", query, err))
	}
	rawJson, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error with ioReadall", err))
	}

	var res ForecastRequest

	err = json.Unmarshal([]byte(rawJson), &res)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error unmarshallling json %s, Json\n%s", err, rawJson))
	}

	//the query is only 4 days
	frostDays := make([]string, 4)
	// check for close to freezing temp
	//based on units
	for _, v := range res.List {
		if v.Main.TempMin < 274 {
			frostDays = append(frostDays, v.Main.Dt_txt)
		}
	}
	if len(frostDays) == 0 {
		return &Response{
			StatusCode: 200,
			Body:       fmt.Sprintf("No frost is expected in %f, %f over the next 4 days!", lat, lon),
		}, nil
	} else {
		return &Response{
			StatusCode: 200,
			Body:       fmt.Sprintf("Frost Warning on the following dates for %f, %f!\n\n: %s", lat, lon, frostDays),
		}, nil
	}
}
