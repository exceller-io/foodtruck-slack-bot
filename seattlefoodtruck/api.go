package seattlefoodtruck

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//NeighborhoodResponse neighborhood api response
type NeighborhoodResponse struct {
	Pagination struct {
		Page       int `json:"page"`
		TotalPages int `json:"total_pages"`
		TotalCount int `json:"total_count"`
	} `json:"pagination"`
	Neighborhoods []struct {
		Name        string `json:"name"`
		Latitude    string `json:"latitude"`
		Longitude   string `json:"longitude"`
		Description string `json:"description"`
		ZoomLevel   int    `json:"zoom_level"`
		Photo       string `json:"photo"`
		ID          string `json:"id"`
		UID         int    `json:"uid"`
	} `json:"neighborhoods"`
}

//LocationRequest request to get locations
type LocationRequest struct {
	Page         int
	Neighborhood string
}

func (lr LocationRequest) toQueryString() string {
	var qs = "?"

	if len(lr.Neighborhood) == 0 {
		lr.Neighborhood = "bellevue" //set default to bellevue
	}
	qs += "page=" + strconv.Itoa(lr.Page) + "&only_with_events=true" + "&neighborhood=" + lr.Neighborhood +
		"&with_active_trucks=true"

	return strings.TrimSpace(qs)
}

//LocationResponse Get locations response
type LocationResponse struct {
	Pagination struct {
		Page       int `json:"page"`
		TotalPages int `json:"total_pages"`
		TotalCount int `json:"total_count"`
	} `json:"pagination"`
	Locations []struct {
		Name            string  `json:"name"`
		Longitude       float64 `json:"longitude"`
		Latitude        float64 `json:"latitude"`
		Address         string  `json:"address"`
		Photo           string  `json:"photo"`
		GooglePlaceID   string  `json:"google_place_id"`
		CreatedAt       string  `json:"created_at"`
		NeighborhoodID  int     `json:"neighborhood_id"`
		Slug            string  `json:"slug"`
		FilteredAddress string  `json:"filtered_address"`
		ID              string  `json:"id"`
		UID             int     `json:"uid"`
		Neighborhood    struct {
			Name string `json:"name"`
			ID   int    `json:"id"`
		} `json:"neighborhood"`
		Pod struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"pod,omitempty"`
	} `json:"locations"`
}

//LocationEventsRequest is request for seattle food trucks API at a location
type LocationEventsRequest struct {
	Location int
	Page     int
}

//NewLocationEventsRequest returns a new request initialized
func NewLocationEventsRequest(location int, page int) LocationEventsRequest {
	return LocationEventsRequest{
		Location: location,
		Page:     page,
	}
}

//FoodTruck Booked food truck
type FoodTruck struct {
	Name           string   `json:"name"`
	Trailer        bool     `json:"trailer"`
	FoodCategories []string `json:"food_categories"`
	ID             string   `json:"id"`
	UID            int      `json:"uid"`
	FeaturedPhoto  string   `json:"featured_photo"`
}

//Booking Event Booking
type Booking struct {
	ID     int       `json:"id"`
	Status string    `json:"status"`
	Paid   bool      `json:"paid"`
	Truck  FoodTruck `json:"truck"`
}

//Event Event at a location
type Event struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	EventID     int       `json:"event_id"`
	Bookings    []Booking `json:"bookings"`
}

//Pagination pagination info
type Pagination struct {
	Page       int `json:"page"`
	TotalPages int `json:"total_pages"`
	TotalCount int `json:"total_count"`
}

//LocationEventsResponse is response from Seattle Food Trucks API
type LocationEventsResponse struct {
	Paging Pagination `json:"pagination"`
	Events []Event    `json:"events"`
}

func (ler LocationEventsRequest) toQueryString() string {
	var qs = "?"
	if ler.Location == 0 {
		ler.Location = 44 //set it to T-Mobile factoria location as default
	}
	if ler.Page == 0 {
		ler.Page = 1
	}
	qs += "page=" + strconv.Itoa(ler.Page) + "&for_locations=" + strconv.Itoa(ler.Location) + "&with_active_trucks=true" +
		"&include_bookings=true" + "&with_booking_status=approved"

	return qs
}

//Proxy is a Seattle Food Truck API proxy
type Proxy struct {
	HTTPClient *http.Client
	BaseURL    string
}

//NewProxy creates a new proxy
func NewProxy(baseURL string) (Proxy, error) {
	var p Proxy

	if len(baseURL) == 0 {
		return p, fmt.Errorf("Invalid Parameter: baseURL is missing")
	}
	p = Proxy{
		HTTPClient: http.DefaultClient,
		BaseURL:    baseURL,
	}
	return p, nil
}

//GetLocationEvents gets events for a specific location
func (p Proxy) GetLocationEvents(request *LocationEventsRequest) (LocationEventsResponse, error) {
	var r LocationEventsResponse
	var api = "/api/events"

	if request == nil {
		return r, fmt.Errorf("Invalid Request")
	}
	//convert the request into a querystring and concatenate to API string
	qs := request.toQueryString()
	api += qs

	httpRequest, err := http.NewRequest("GET", p.BaseURL+api, nil)
	if err != nil {
		return r, fmt.Errorf("An error occured creating http request")
	}
	//Execute the request
	httpResponse, err := p.HTTPClient.Do(httpRequest)
	if err != nil {
		return r, fmt.Errorf("An Error occurred querying location events using seattle food trucks api")
	}
	//Response body must be closed
	defer httpResponse.Body.Close()
	if err := json.NewDecoder(httpResponse.Body).Decode(&r); err != nil {
		log.Println(err)
		return r, fmt.Errorf(err.Error())
	}
	return r, nil
}

//GetNeighborhoods gets seattle neighborhoods where you can find food trucks
func (p Proxy) GetNeighborhoods() (NeighborhoodResponse, error) {
	var nr NeighborhoodResponse
	var api = "/api/neighborhoods"

	httpRequest, err := http.NewRequest("GET", p.BaseURL+api, nil)
	if err != nil {
		return nr, fmt.Errorf("An error occurred creating http request")
	}
	//Execute the request
	httpResponse, err := p.HTTPClient.Do(httpRequest)
	if err != nil {
		return nr, fmt.Errorf("An error occurred querying neighborhoods using seattle food trucks api")
	}
	//Response body must be closed
	defer httpResponse.Body.Close()
	if err := json.NewDecoder(httpResponse.Body).Decode(&nr); err != nil {
		return nr, fmt.Errorf(err.Error())
	}
	return nr, nil
}

//GetLocations gets all locations at a specific neighborhood in seattle where you can find food trucks
func (p Proxy) GetLocations(request *LocationRequest) (LocationResponse, error) {
	var lr LocationResponse
	var api = "/api/locations"

	if request == nil {
		return lr, fmt.Errorf("Invalid Request")
	}
	qs := request.toQueryString()
	api += qs
	httpRequest, err := http.NewRequest("GET", p.BaseURL+api, nil)
	if err != nil {
		return lr, fmt.Errorf("An error occured creating http request")
	}
	//Execute the request
	httpResponse, err := p.HTTPClient.Do(httpRequest)
	if err != nil {
		return lr, fmt.Errorf("An Error occurred querying location using seattle food trucks api")
	}
	//Response body must be closed
	defer httpResponse.Body.Close()
	if err := json.NewDecoder(httpResponse.Body).Decode(&lr); err != nil {
		log.Println(err)
		return lr, fmt.Errorf(err.Error())
	}
	return lr, nil
}
