package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// router is defined as a package variable because it doesn't seem likely that anyone importing
// this package will want to create multiple routers with the same behavior
var router *mux.Router

// supportedOperations defines a list of accepted endpoints and the associated math operations
var supportedOperations map[string]func(float64, float64) float64 = map[string]func(float64, float64) float64{
	"add":      func(x, y float64) float64 { return x + y },
	"subtract": func(x, y float64) float64 { return x - y },
	"multiply": func(x, y float64) float64 { return x * y },
	"divide":   func(x, y float64) float64 { return x / y },
}

func init() {
	router = mux.NewRouter()
	router.HandleFunc("/{op}", mathHandler)
}

// GetRouter just returns this package's http router
func GetRouter() *mux.Router {
	return router
}

// mathHandler parses two arguments 'x' and 'y' from the client, applies the requested math operation,
// builds a MathOKResponse struct, JSON encodes it, and returns it
func mathHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // just in case

	muxVars := mux.Vars(r)
	op := muxVars["op"]

	x, y, err := parseClientVars(r)
	if err != nil {
		status, resBytes := createErrorResponse(http.StatusBadRequest, err)
		w.WriteHeader(status)
		w.Write(resBytes)
		return
	}

	if supportedOperations[op] == nil {
		status, resBytes := createErrorResponse(http.StatusBadRequest, err)
		w.WriteHeader(status)
		w.Write(resBytes)
		return
	}

	answer, inCache := RetrieveFromCache(op, x, y)
	if !inCache {
		answer = supportedOperations[op](x, y)
	}

	AddToCache(op, x, y, answer)

	okResponse := MathOKResponse{
		Action: op,
		X:      x,
		Y:      y,
		Answer: answer,
		Cached: inCache,
	}
	okResBytes, err := json.Marshal(okResponse)
	if err != nil {
		status, resBytes := createErrorResponse(http.StatusInternalServerError, err)
		w.WriteHeader(status)
		w.Write(resBytes)
		return
	}

	w.WriteHeader(http.StatusOK) // don't need to call for 200 OK, but I prefer to be explicit
	w.Write(okResBytes)
}

// createErrorResponse attempts to build a MathErrorResponse based upon the provided status and error.
// If there's an error marshalling the object, it returns a 500 Internal Server Error and an empty
// body
func createErrorResponse(status int, e error) (int, []byte) {
	errResponse := MathErrorResponse{
		Status: http.StatusBadRequest,
		Error:  e.Error(),
		// for simplicity's sake, we're trusting the client with the content of our error
	}

	resBytes, err := json.Marshal(errResponse)
	if err != nil {
		return http.StatusInternalServerError, nil
	}

	return http.StatusBadRequest, resBytes
}
