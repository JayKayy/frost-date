package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func Main(req Request) *Response {
	// https:openweathermap.org/api/one-call-api#parameter

	if req.Lat == "" || req.Lon == "" {
		return &Response{
			Body: "Lat and Lon parameters are required",
		}
	}

	lat, errla := strconv.ParseFloat(req.Lat, 64)
	lon, errlon := strconv.ParseFloat(req.Lon, 64)

	// Set API token in serverless Env
	apiKey := os.Getenv("APIKEY")
	if apiKey == "" {
		return &Response{
			Body: "No APIKEY detected",
		}
	}
	if errla != nil || errlon != nil {
		//	return nil, fmt.Errorf("Error parsing lat and lon as floats")
		return &Response{
			Body: "Error parsing latitude and longitude as floats",
		}
	}

	query := fmt.Sprintf("https://api.openweathermap.org/data/2.5/onecall?lat=%f&lon=%f&appid=%s&exclude=current,minutely,hourly,alerts", lat, lon, apiKey)
	r, err := http.Get(query)
	rawJson, _ := io.ReadAll(r.Body)
	if r.StatusCode != 200 {
		//return nil, errors.New(fmt.Sprint("Error with ioReadall", err))
		return &Response{
			Body: fmt.Sprintf("Error from query! %d\n%s\n", r.StatusCode, string(rawJson)),
		}
	}
	defer r.Body.Close()

	var res Forecast

	err = json.Unmarshal([]byte(rawJson), &res)
	if err != nil {
		return &Response{
			Body: fmt.Sprintf("Error unable to Marchall response. %s", err),
		}
	}
	var frostDays string
	// check for close to freezing temp in Kelvin
	for _, day := range res.Daily {
		if day.Temp.Min < 274 {
			frostDays = fmt.Sprintf("%s\n", time.Unix(day.Dt, 0).String())
		}
	}

	if len(frostDays) == 0 {
		return &Response{
			StatusCode: 200,
			Body:       fmt.Sprintf("No frost is expected in %f, %f over the next 7 days!", lat, lon),
		}

	} else {
		frostDays = strings.TrimSuffix(frostDays, "\n")
		return &Response{
			StatusCode: 200,
			Body:       fmt.Sprintf("Frost Warning on the following dates for %f, %f!\n %s", lat, lon, frostDays),
		}
	}
}
