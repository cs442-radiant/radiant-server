package main

import "net/http"

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"GetRestaurant",
		"GET",
		"/restaurant/{restaurantName}",
		GetRestaurant,
	},
	Route{
		"PostBundle",
		"POST",
		"/bundle",
		PostBundle,
	},
	Route{
		"PostSample",
		"POST",
		"/sample",
		PostSample,
	},
	Route{
		"PostLearn",
		"POST",
		"/learn",
		PostLearn,
	},
	Route{
		"PostCurrentLocation",
		"POST",
		"/location",
		PostCurrentLocation,
	},
}
