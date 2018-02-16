package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
)

// acceptedContentTypes maps content-types that we've written parsing logic for to the functions
// that perform that parsing. This is also an O(1) way of checking if we support a given content-type
var acceptedContentTypes = map[string]func(*http.Request) (float64, float64, error){
	"application/json":                  parseJSON,
	"application/x-www-form-urlencoded": parseFormURLEncoded,
}

// parseClientVars attempts to determine the request's content-type and parse the
// variables 'x' and 'y' accordingly
func parseClientVars(r *http.Request) (float64, float64, error) {
	contentType := r.Header.Get("content-type")
	if contentType == "" {
		return 0, 0, fmt.Errorf("no content-type specified")
	}

	if acceptedContentTypes[contentType] == nil {
		return 0, 0, fmt.Errorf("unsupported content-type: %q", contentType)
	}

	return acceptedContentTypes[contentType](r)
}

// parseJSON attempts to decode the request body into JSON and then returns
// the X and Y fields (see MathRequest)
func parseJSON(r *http.Request) (float64, float64, error) {
	var mathReq MathRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&mathReq)
	if err != nil {
		return 0, 0, errors.Wrap(err, "json decode error")
	}

	return mathReq.X, mathReq.Y, nil
}

// parseFormURLEncoded parses the request form and returns the form values 'x' and 'y'
func parseFormURLEncoded(r *http.Request) (float64, float64, error) {
	err := r.ParseForm()
	if err != nil {
		return 0, 0, errors.Wrap(err, "parse request form failed")
	}

	xStr := r.Form.Get("x")
	x, err := strconv.ParseFloat(xStr, 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "parse x failed")
	}

	yStr := r.Form.Get("y")
	y, err := strconv.ParseFloat(yStr, 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "parse y failed")
	}

	return x, y, nil
}
