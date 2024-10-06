package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

type GeocodeResponse []struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	State   string  `json:"state,omitempty"`
	Country string  `json:"country"`
}

type OneCallResponse struct {
	Current struct {
		Temp    float64 `json:"temp"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	} `json:"current"`
	Daily []struct {
		Temp struct {
			Day float64 `json:"day"`
		} `json:"temp"`
		Weather []struct {
			Description string `json:"description"`
		} `json:"weather"`
	} `json:"daily"`
}

var apiKey string

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		os.Exit(1)
	}

	apiKey = os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set OPENWEATHER_API_KEY in .env file")
		os.Exit(1)
	}

	if len(os.Args) < 3 {
		fmt.Println("Please provide a city name and state/country")
		os.Exit(1)
	}
	city := os.Args[1]
	region := os.Args[2]

	lat, lon, location, err := getCoordinates(city, region)
	if err != nil {
		fmt.Println("Error fetching coordinates:", err)
		os.Exit(1)
	}

	fmt.Printf("Fetching weather for: %s\n", location)
	getWeather(lat, lon)
}

func getCoordinates(city, region string) (float64, float64, string, error) {
	query := fmt.Sprintf("%s,%s", city, region)
	apiUrl := fmt.Sprintf("http://api.openweathermap.org/geo/1.0/direct?q=%s&limit=1&appid=%s", query, apiKey)
	resp, err := http.Get(apiUrl)
	if err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, "", fmt.Errorf("unable to fetch coordinates for city %s", city)
	}

	var data GeocodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, 0, "", err
	}

	if len(data) == 0 {
		return 0, 0, "", fmt.Errorf("city %s not found", city)
	}

	// Format the location
	location := data[0].Name
	if data[0].State != "" {
		location += ", " + data[0].State
	} else {
		location += ", " + data[0].Country
	}

	return data[0].Lat, data[0].Lon, location, nil
}

func getWeather(lat, lon float64) {
	apiUrl := fmt.Sprintf("https://api.openweathermap.org/data/3.0/onecall?lat=%f&lon=%f&exclude=minutely,hourly,alerts&units=imperial&appid=%s", lat, lon, apiKey)
	resp, err := http.Get(apiUrl)

	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error: Unable to fetch weather data!")
		os.Exit(1)
	}

	var data OneCallResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		fmt.Println("Error decoding JSON:", err)
		os.Exit(1)
	}

	fmt.Printf("Current Weather: %.1f°F, %s\n", data.Current.Temp, data.Current.Weather[0].Description)
	fmt.Println("7-Day Forecast:")
	for i, day := range data.Daily {
		if i == 0 {
			continue // Skip the current day (already displayed)
		}
		fmt.Printf("Day %d: %.1f°F, %s\n", i, day.Temp.Day, day.Weather[0].Description)
	}
}
