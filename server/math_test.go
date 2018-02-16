package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func init() {
	// don't want standard log output during testing because we're expecting erroneous requests
	log.SetOutput(ioutil.Discard)
}

// TestMathHandler makes a series of requests to mathHandler and checks results for proper format,
// answer accuracy, and "cache" field accuracy
func TestMathHandler(t *testing.T) {
	cleanUpCache()
	t.Run("form url encoded", formURLEncodedRequest)
	cleanUpCache()
	t.Run("json encoded", jsonRequest)
	cleanUpCache()
}

// formURLEncodedRequest tests a variety of requests with content-type application/x-www-form-urlencoded
// Despite what I stated concering the value of repeated code in some of the other test files, repeated
// code would quickly become tedious in this case, so we're looping over supportedOperations
func formURLEncodedRequest(t *testing.T) {
	contentType := "application/x-www-form-urlencoded"

	for operation := range supportedOperations {
		t.Log(operation)
		expectedX, expectedY := 34.854, -0.935
		if math.IsNaN(supportedOperations[operation](expectedX, expectedY)) {
			// make y more well-behaved for pow, root, and log
			expectedY = 1.20034
		}

		reqURL := fmt.Sprintf("http://localhost:8080/%s?x=%f&y=%f", operation, expectedX, expectedY)
		// FIXME: method doesn't currently matter (we're not checking it), but this will need to change
		// if we begin checking method
		req := httptest.NewRequest(http.MethodPost, reqURL, nil)
		req.Header.Set("Content-Type", contentType)

		cleanUpCache()                                               // not sure about best practice on borrowing this from cache_test.go
		validRequest(t, operation, expectedX, expectedY, false, req) // first request w/o cached response
		validRequest(t, operation, expectedX, expectedY, true, req)  // second expects cached response

		// this function waits the full answer expiration time for each operation
		if !testing.Short() {
			originalExpiration := cacheExpiration
			cacheExpiration = time.Millisecond * 50

			cleanUpCache()
			validRequest(t, operation, expectedX, expectedY, false, req)

			time.Sleep(cacheExpiration)
			validRequest(t, operation, expectedX, expectedY, false, req)

			cacheExpiration = originalExpiration
		}

		// missing content type
		noContentTypeReq := httptest.NewRequest(http.MethodPost, reqURL, nil)
		errorRequest(t, http.StatusBadRequest, noContentTypeReq)

		// unsupported operation
		unsupportedOpURL := fmt.Sprintf("http://localhost:8080/fourierTransform?x=1.0&y=-1.0")
		unsupportedOpReq := httptest.NewRequest(http.MethodPost, unsupportedOpURL, nil)
		unsupportedOpReq.Header.Set("Content-Type", contentType)
		errorRequest(t, http.StatusBadRequest, unsupportedOpReq)
	}
}

// jsonRequest performs the same set of tests as formURLEncodedRequest, but with a JSON encoded request body
func jsonRequest(t *testing.T) {
	contentType := "application/json"

	for operation := range supportedOperations {
		t.Log(operation)
		expectedX, expectedY := -44.444, 1.000001
		if math.IsNaN(supportedOperations[operation](expectedX, expectedY)) {
			// make x and y more well-behaved for pow, root, and log
			expectedX, expectedY = 26.8834, 7.00849
		}

		reqURL := fmt.Sprintf("http://localhost:8000/%s", operation)

		mathReq := MathRequest{
			X: expectedX,
			Y: expectedY,
		}
		bodyBytes, err := json.Marshal(mathReq)
		if err != nil {
			t.Fatalf("json marshal failed: %s\n", err)
		}

		req := httptest.NewRequest(http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", contentType)

		cleanUpCache()
		validRequest(t, operation, expectedX, expectedY, false, req)

		// easier than doing type assertion on req.Body then calling Reset()
		req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
		validRequest(t, operation, expectedX, expectedY, true, req)

		if !testing.Short() {
			originalExpiration := cacheExpiration
			cacheExpiration = time.Millisecond * 50
			cleanUpCache()

			req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			validRequest(t, operation, expectedX, expectedY, false, req)

			time.Sleep(cacheExpiration)
			req.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
			validRequest(t, operation, expectedX, expectedY, false, req)

			cacheExpiration = originalExpiration
		}

		// missing content type
		noContentTypeReq := httptest.NewRequest(http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
		errorRequest(t, http.StatusBadRequest, noContentTypeReq)

		// unsupported operation
		unsupportedOpURL := fmt.Sprintf("http://localhost:8080/gradientDescent")
		unsupportedOpReq := httptest.NewRequest(http.MethodPost, unsupportedOpURL, bytes.NewReader(bodyBytes))
		unsupportedOpReq.Header.Set("Content-Type", contentType)
		errorRequest(t, http.StatusBadRequest, unsupportedOpReq)
	}
}

// validRequest makes a correctly formatted request to the router and checks the response for errors.
func validRequest(t *testing.T, expectedOp string, expectedX, expectedY float64, expectedCachedVal bool, req *http.Request) {
	// can only create expectedAns this way because this test checks valid requests only
	expectedAns := supportedOperations[expectedOp](expectedX, expectedY)
	resRecorder := httptest.NewRecorder()

	GetRouter().ServeHTTP(resRecorder, req)

	var mathRes MathOKResponse
	decoder := json.NewDecoder(resRecorder.Body)
	err := decoder.Decode(&mathRes)
	if err != nil {
		t.Fatalf("json decode failed: %s\n", err)
	}

	if mathRes.Action != expectedOp {
		t.Logf("unexpected action value: (actual %s != expected %s)\n", mathRes.Action, expectedOp)
		t.Fail()
	}

	if mathRes.X != expectedX {
		t.Logf("unexpected x value: (actual %f != expected %f)\n", mathRes.X, expectedX)
		t.Fail()
	}

	if mathRes.Y != expectedY {
		t.Logf("unexpected y value: (actual %f != expected %f)\n", mathRes.Y, expectedY)
		t.Fail()
	}

	if mathRes.Answer != expectedAns {
		t.Logf("unexpected ans value: (actual %f != expected %f)\n", mathRes.Answer, expectedAns)
		t.Fail()
	}

	if mathRes.Cached != expectedCachedVal {
		t.Logf("unexpected cached value: (actual %t != expected %t)\n", mathRes.Cached, expectedCachedVal)
		t.Fail()
	}
}

// errorRequest makes an incorrectly formatted request to the router and checks the returned status as
// well as the "status" field of the returned JSON (the only time mathHandler doesn't return JSON is
// when there's a JSON marshalling error, which is very infrequent given what we're marhsalling)
func errorRequest(t *testing.T, expectedStatus int, req *http.Request) {
	resRecorder := httptest.NewRecorder()

	GetRouter().ServeHTTP(resRecorder, req)

	if resRecorder.Code != expectedStatus {
		t.Logf("unexpected status value: (actual %d != expected %d)\n", resRecorder.Code, expectedStatus)
		t.Fail()
	}

	var errRes MathErrorResponse
	decoder := json.NewDecoder(resRecorder.Body)
	err := decoder.Decode(&errRes)
	if err != nil {
		t.Fatalf("json decode failed: %s\n", err)
	}

	if errRes.Status != expectedStatus {
		t.Logf("unexpected status value: (actual %d != expected %d)\n", errRes.Status, expectedStatus)
		t.Fail()
	}
}
