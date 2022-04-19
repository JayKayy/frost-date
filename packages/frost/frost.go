package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
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
type Forecast struct {
	Lat   float64
	Lon   float64
	Daily []Day `json:"daily,omitempty"`
}

type Day struct {
	Dt   int64
	Temp Temp
}
type Temp struct {
	Min float64
	Max float64
}

// Help from https://www.omnicalculator.com/other/mayan-calendar#how-to-convert-a-date-to-the-long-count-calendar

// Calculators tools
// https://utahgeology.com/bin/maya-calendar-converter/
// https://maya.nmai.si.edu/calendar/maya-calendar-converter

//func Main(req Request) (*Response, error) {
func main() {

	req := Request{Lat: "19.39068", Lon: "-99.2836969"}
	//hash of month and day lengths for validation
	//	monthLengths := map[int]int{0: 31, 1: 29, 2: 31, 3: 30, 4: 31, 5: 30, 6: 31, 7: 31, 8: 30, 9: 31, 10: 30, 11: 31}
	//	months := map[int]string{0: "January", 1: "February", 2: "March", 3: "April", 4: "May", 5: "June", 6: "July", 7: "August", 8: "September", 9: "October", 10: "November", 11: "December"}

	// Used for local testing
	if req.Lat == "" || req.Lon == "" {
		//return nil, errors.New("Lat and Lon parameters are required")
		fmt.Println("error lat and long need to be specified")
		return
	}

	lat, errla := strconv.ParseFloat(req.Lat, 64)
	lon, errlon := strconv.ParseFloat(req.Lon, 64)

	// Set API token in serverless Env
	apiKey := os.Getenv("APIKEY")
	if apiKey == "" {
		fmt.Println("No API key Set!")
		return
	}
	if errla != nil || errlon != nil {
		//	return nil, fmt.Errorf("Error parsing lat and lon as floats")
		fmt.Println("Error parsing lat and lon as floats")
		return
	}

	query := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall?lat=%f&lon=%f&appid=%s&exclude=current,minutely,hourly,alerts", lat, lon, apiKey)
	r, err := http.Get(query)
	rawJson, _ := io.ReadAll(r.Body)
	if r.StatusCode != 200 {
		//return nil, errors.New(fmt.Sprint("Error with ioReadall", err))
		fmt.Printf("error from query! %d\n%s\n", r.StatusCode, string(rawJson))
		return
	}
	defer r.Body.Close()

	var res Forecast

	err = json.Unmarshal([]byte(rawJson), &res)
	if err != nil {
		//return nil, errors.New(fmt.Sprintf("Error unmarshallling json %s, Json\n%s", err, rawJson))
		fmt.Println("error from unmarshall", err)
		return
	}
	var frostDays []time.Time
	// check for close to freezing temp
	//based on units
	for _, day := range res.Daily {
		if day.Temp.Min < 274 {
			frostDays = append(frostDays, time.Unix(day.Dt, 0))
		}
	}
	if len(frostDays) == 0 {
		// return &Response{
		// 	StatusCode: 200,
		// 	Body:       fmt.Sprintf("No frost is expected in %f, %f over the next 4 days!", lat, lon),
		// }, nil
		fmt.Println("No Frost!")

	} else {
		// return &Response{
		// 	StatusCode: 200,
		// 	Body:       fmt.Sprintf("Frost Warning on the following dates for %f, %f! %s", lat, lon, frostDays),
		// }, nil

		fmt.Println("Frost dates: \n ", frostDays)
	}
}
