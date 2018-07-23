package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

// Insert your ipstack and wunderground API keys here
var ipstack_key = "INSERT KEY HERE"
var wunderground_key = "INSERT KEY HERE"

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

// HTTP handler
func handler(w http.ResponseWriter, r *http.Request) {
	// Get users ip
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	// Locate user
	geolocation := getgeolocationdata(ip)
	// Get weather data for location
	data := getweatherdata(geolocation)
	// Serve template
	tmpl := template.Must(template.ParseFiles("template.html"))
	tmpl.Execute(w, data)
}

// Geolocation struct
type geolocationdata struct {
	Ip      string
	Country string
	City    string
}

// Get user geolocation from ip
func getgeolocationdata(ip string) geolocationdata {
	// Send request to API
	r, err := http.Get("http://api.ipstack.com/" + ip + "?output=json&access_key=" + ipstack_key)
	checkErr(err)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	// Extract data from response and return it
	// If in US
	if gjson.Get(string(body), "country_code").String() == "US" {
		// Return state instead of country
		return geolocationdata{Ip: ip, Country: gjson.Get(string(body), "region_code").String(), City: gjson.Get(string(body), "city").String()}
	}
	return geolocationdata{Ip: ip, Country: gjson.Get(string(body), "country_code").String(), City: gjson.Get(string(body), "city").String()}
}

// Weather struct
type weatherdata struct {
	Location       string
	Currentweather string
	Currenttemp    string
	Dailyweather   []dailyweatherdata
}
type dailyweatherdata struct {
	Day     string
	Weather string
	High    string
	Low     string
}

// Get 3 day forecast from wunderground
func getweatherdata(geolocation geolocationdata) weatherdata {
	// Send request to API
	r, err := http.Get("http://api.wunderground.com/api/" + wunderground_key + "/conditions/forecast/q/" + geolocation.Country + "/" + geolocation.City + ".json")
	checkErr(err)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	// Find what days are being forecasted and convert them into 3 character snippets
	var days []string
	for _, ds := range gjson.Get(string(body), "forecast.txt_forecast.forecastday.#.title").Array() {
		if strings.Contains(ds.String(), "Night") == false {
			days = append(days, ds.String()[0:3])
		}
	}
	// Converts weather to an icon
	makeicon := func(weather string) string {
		switch weather {
		case "sunny":
			return "0"
		case "rain":
			return "3"
		case "clear":
			return "0"
		case "sleet":
			return "r"
		case "cloudy":
			return "2"
		case "snow":
			return "s"
		case "flurries":
			return "r"
		case "fog":
			return "f"
		case "hazy":
			return "k"
		case "chanceflurries":
			return "v"
		case "chancerain":
			return "d"
		case "chancetstorms":
			return "x"
		case "chancesleet":
			return "v"
		case "chancesnow":
			return "n"
		case "partlycloudy":
			return "3"
		case "partlysunny":
			return "3"
		case "mostlycloudy":
			return "3"
		case "tstorms":
			return "t"
		default:
			return "0"
		}
	}
	// Weather for each day
	var weathers []string
	for _, ws := range gjson.Get(string(body), "forecast.txt_forecast.forecastday.#.icon").Array() {
		if strings.Contains(ws.String(), "nt_") == false {
			// Convert the weather to an icon
			weathers = append(weathers, makeicon(ws.String()))
		}
	}
	// High for each day
	var highs []string
	for _, hs := range gjson.Get(string(body), "forecast.simpleforecast.forecastday.#.high.celsius").Array() {
		highs = append(highs, hs.String())
	}
	// Low for each day
	var lows []string
	for _, ls := range gjson.Get(string(body), "forecast.simpleforecast.forecastday.#.low.celsius").Array() {
		lows = append(lows, ls.String())
	}
	// Combine data for each day
	var dailyweatherdatas []dailyweatherdata
	for i, day := range days {
		dailyweatherdatas = append(dailyweatherdatas, dailyweatherdata{Day: day, Weather: weathers[i], High: highs[i], Low: lows[i]})
	}
	return weatherdata{Location: strings.Replace(geolocation.City, "-", " ", -1), Currentweather: makeicon(gjson.Get(string(body), "current_observation.icon").String()), Currenttemp: strconv.FormatFloat(gjson.Get(string(body), "current_observation.temp_c").Float(), 'f', 0, 64), Dailyweather: dailyweatherdatas}
}

// Error handler
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
