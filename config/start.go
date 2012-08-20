package config

import (
	"appengine"
	"net/http"
)

// Initialize is call on the first request a new instance recieves. Any
// code that need to run prior to the first request should go here.
func Start(r *http.Request) {
	c := appengine.NewContext(r)
	c.Infof(`config: New instance starting...`)
}
