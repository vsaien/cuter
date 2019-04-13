package cuter

import "net/http"

type (
	Route struct {
		Method  string
		Path    string
		Handler http.HandlerFunc
	}
	featuredRoutes struct {
		routes []Route
	}
	RouteOption func(r *featuredRoutes)
)
