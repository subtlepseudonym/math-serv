package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"github.com/gorilla/mux"
)

// router is defined as a package variable because it doesn't seem likely that anyone importing
// this package will want to create multiple routers with the same behavior
var router *mux.Router

// supportedOperations defines a list of accepted endpoints and the associated math operations.
// I'm a big supporter of maps of functions. They increase lookup time (on the part of anyone reading
// the code), but they can considerably decrease code repetition and make extensibility easy
var supportedOperations = map[string]func(float64, float64) float64{
	"add":      func(x, y float64) float64 { return x + y },
	"subtract": func(x, y float64) float64 { return x - y },
	"multiply": func(x, y float64) float64 { return x * y },
	"divide":   func(x, y float64) float64 { return x / y },
	"mod":      func(x, y float64) float64 { return math.Mod(x, y) },
	"pow":      func(x, y float64) float64 { return math.Pow(x, y) },
	"root":     func(x, y float64) float64 { return math.Pow(x, 1/y) },
	"log":      func(x, y float64) float64 { return math.Log(x) / math.Log(y) },
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
// builds a MathOKResponse struct, JSON encodes it, and returns it.
// This functions sets off gocyclo for cyclomatic complexity (11), but I'm going to let it go considering
// it's the only handler
func mathHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r.Body == nil {
			return
		}
		err := r.Body.Close() // just in case
		if err != nil {
			log.Printf("req body close failed: %s\n", err)
		}
	}()

	muxVars := mux.Vars(r)
	op := muxVars["op"]

	x, y, err := parseClientVars(r)
	if err != nil {
		log.Printf("parse client vars failed: %s\n", err)
		status, resBytes := createErrorResponse(http.StatusBadRequest, err)
		w.WriteHeader(status)
		_, err = w.Write(resBytes)
		if err != nil {
			// bummer, most we can do is log the error
			log.Printf("response write failed: %s\n", err)
		}
		return
	}

	if supportedOperations[op] == nil {
		errStr := fmt.Sprintf("unsupported operation request: %q", op)
		log.Printf(errStr)
		status, resBytes := createErrorResponse(http.StatusBadRequest, fmt.Errorf(errStr))
		w.WriteHeader(status)
		_, err = w.Write(resBytes)
		if err != nil {
			log.Printf("response write failed: %s\n", err)
		}
		return
	}

	answer, inCache := retrieveFromCache(op, x, y)
	if !inCache {
		answer = supportedOperations[op](x, y)
	}

	addToCache(op, x, y, answer)

	okResponse := MathOKResponse{
		Action: op,
		X:      x,
		Y:      y,
		Answer: answer,
		Cached: inCache,
	}
	okResBytes, err := json.Marshal(okResponse)
	if err != nil {
		// included mathHandler in error log because we have the same error log description
		// in createErrorResponse
		log.Printf("mathHandler: json marshal failed: %s\n", err)
		status, resBytes := createErrorResponse(http.StatusInternalServerError, err)
		w.WriteHeader(status)
		_, err = w.Write(resBytes)
		if err != nil {
			log.Printf("response write failed: %s\n", err)
		}
		return
	}

	w.WriteHeader(http.StatusOK) // don't need to call for 200 OK, but I prefer to be explicit
	_, err = w.Write(okResBytes)
	if err != nil {
		log.Printf("response write failed: %s\n", err)
	}
}

// createErrorResponse attempts to build a MathErrorResponse based upon the provided status and error.
// If there's an error marshalling the object, it returns a 500 Internal Server Error and an empty
// body.
// For errors, I'm attempting to send a representative JSON object back to the client, but that obviously
// opens us up to json.Marshal() errors.  Not entirely sure what best practice is for returning errors to
// the client, so I'm assuming JSON because that's the content-type that proper responses return in
func createErrorResponse(status int, e error) (int, []byte) {
	errResponse := MathErrorResponse{
		Status: http.StatusBadRequest,
		Error:  e.Error(),
		// for simplicity's sake, we're trusting the client with the content of our error
	}

	resBytes, err := json.Marshal(errResponse)
	if err != nil {
		log.Printf("createErrorResponse: json marshal failed: %s\n", err)
		return http.StatusInternalServerError, nil
	}

	return http.StatusBadRequest, resBytes
}
