package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Normally I'm very against copy / pasting code, but Mitchell Hashimoto's talk at GopherCon 2017
// has convinced me that it may be apropriate during testing for clarity

// TestParseClientVars attempts to test all the cases in which parseClientVars will be used.
// parseJSON and parseFormURLEncoded are not tested explicitly because they are called as a part
// of parseClientVars execution
func TestParseClientVars(t *testing.T) {
	t.Run("form with header", parseFormWithHeader)
	t.Run("form sans header", parseFormWithoutHeader)
	t.Run("json with header", parseJSONWithHeader)
	t.Run("json sans header", parseJSONWithoutHeader)
	t.Run("unsupported type", parseUnsupportedContentType)
}

func parseFormWithHeader(t *testing.T) {
	expectedX, expectedY := 7.0, 8.0
	reqURLStr := fmt.Sprintf("http://localhost:8080/multiply?x=%f&y=%f", expectedX, expectedY)

	req := httptest.NewRequest(http.MethodGet, reqURLStr, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded") // providing content-type header

	actualX, actualY, err := parseClientVars(req)
	if err != nil {
		// not expecting an error here
		t.Logf("unexpected error: %s\n", err)
		t.Fail()
	}

	if actualX != expectedX {
		t.Logf("X value mismatch: (actual %f != expected %f )\n", actualX, expectedX)
		t.Fail()
	}
	if actualY != expectedY {
		t.Logf("Y value mismatch: (actual %f != expected %f)\n", actualY, expectedY)
		t.Fail()
	}
}

func parseFormWithoutHeader(t *testing.T) {
	expectedX, expectedY := 0.0, 0.0
	providedX, providedY := 4.2, 12.6678
	reqURLStr := fmt.Sprintf("http://localhost:8080/divide?x=%f&y=%f", providedX, providedY)

	req := httptest.NewRequest(http.MethodGet, reqURLStr, nil)
	req.Header.Set("Content-Type", "") // no content-type header!

	actualX, actualY, err := parseClientVars(req)
	if err == nil {
		// expecting an error
		t.Log("expecting error, none received")
		t.Fail()
	}

	if actualX != expectedX {
		t.Logf("X value mismatch: (actual %f != expected %f )\n", actualX, expectedX)
		t.Fail()
	}
	if actualY != expectedY {
		t.Logf("Y value mismatch: (actual %f != expected %f)\n", actualY, expectedY)
		t.Fail()
	}
}

func parseJSONWithHeader(t *testing.T) {
	expectedX, expectedY := 9.5334, 2.1
	reqURLStr := "http://localhost:8080/subtract"

	mathReqObj := MathRequest{
		X: expectedX,
		Y: expectedY,
	}

	bodyBytes, err := json.Marshal(mathReqObj)
	if err != nil {
		t.Logf("json marshal failed: %s\n", err)
		t.Fail()
	}

	req := httptest.NewRequest(http.MethodPost, reqURLStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	actualX, actualY, err := parseClientVars(req)
	if err != nil {
		// not expecting an error here
		t.Logf("unexpected error: %s\n", err)
		t.Fail()
	}

	if actualX != expectedX {
		t.Logf("X value mismatch: (actual %f != expected %f )\n", actualX, expectedX)
		t.Fail()
	}
	if actualY != expectedY {
		t.Logf("Y value mismatch: (actual %f != expected %f)\n", actualY, expectedY)
		t.Fail()
	}
}

func parseJSONWithoutHeader(t *testing.T) {
	expectedX, expectedY := 0.0, 0.0
	providedX, providedY := -53.7, 33.2275
	reqURLStr := "http://localhost:8080/add"

	mathReqObj := MathRequest{
		X: providedX,
		Y: providedY,
	}

	bodyBytes, err := json.Marshal(mathReqObj)
	if err != nil {
		t.Logf("json marshal failed: %s\n", err)
		t.Fail()
	}

	req := httptest.NewRequest(http.MethodPost, reqURLStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "")

	actualX, actualY, err := parseClientVars(req)
	if err == nil {
		// expecting an error
		t.Log("expecting error, none received")
		t.Fail()
	}

	if actualX != expectedX {
		t.Logf("X value mismatch: (actual %f != expected %f )\n", actualX, expectedX)
		t.Fail()
	}
	if actualY != expectedY {
		t.Logf("Y value mismatch: (actual %f != expected %f)\n", actualY, expectedY)
		t.Fail()
	}
}

func parseUnsupportedContentType(t *testing.T) {
	expectedX, expectedY := 0.0, 0.0
	providedX, providedY := -53.7, 33.2275
	reqURLStr := "http://localhost:8080/add"

	// FIXME: if we end up supporting xml, this needs to be changed
	bodyBytes := []byte(fmt.Sprintf("<MathRequest><X>%f</X><Y>%f</Y></MathRequest>", providedX, providedY))

	req := httptest.NewRequest(http.MethodPost, reqURLStr, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/xml")

	actualX, actualY, err := parseClientVars(req)
	if err == nil {
		// expecting an error
		t.Log("expecting error, none received")
		t.Fail()
	}

	if actualX != expectedX {
		t.Logf("X value mismatch: (actual %f != expected %f )\n", actualX, expectedX)
		t.Fail()
	}
	if actualY != expectedY {
		t.Logf("Y value mismatch: (actual %f != expected %f)\n", actualY, expectedY)
		t.Fail()
	}
}
