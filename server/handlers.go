package main

import (
	"fmt"
	"net/http"
)

func GetRestaurant(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Get Restaurant!")
}

func GetBundle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Get Bundle!")
}

func PostSample(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Post Sample!")
}
