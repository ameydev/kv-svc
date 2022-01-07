package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a App

func TestMain(m *testing.M) {
	a.Initialize()
	// a.Run()
	code := m.Run()
	os.Exit(code)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestGetEntityBad(t *testing.T) {
	a.InitializeData()
	req, _ := http.NewRequest("GET", "/get/xyz-1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestGetEntityGood(t *testing.T) {
	a.InitializeData()
	req, _ := http.NewRequest("GET", "/get/abc-1", nil)
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var value string
	json.Unmarshal(response.Body.Bytes(), &value)
	if value != "Thermodynamics" {
		t.Errorf(" Expecting an entity with key abc-1 and value Thermodynamics, Got '%s'", value)
	}
}

func TestSetEntityGood(t *testing.T) {
	var jsonStr = []byte(`{"key":"test product", "value": "11.22"}`)
	req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusCreated, response.Code)
}

func TestSetEntityBad(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1","hello": "test"}`)
	req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

func TestSearchPrefixGood(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?prefix=abc-", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestSearchPrefixMultipleKeys(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?prefix=abc", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
	var entities []Entity
	json.Unmarshal(response.Body.Bytes(), &entities)
	if len(entities) != 2 {
		t.Errorf("Expected the 2 keys in the search response but'. Got '%d'", len(entities))
	}
}

func TestSearchPrefixBad(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?prefix=xyz", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestSearchPrefixBadURL(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?prefi", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}

// // searching with suffix
func TestSearchSuffixGood(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?suffix=-1", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusOK, response.Code)
}

func TestSearchSuffixBad(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?suffix=-3", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusNotFound, response.Code)
}

func TestSearchSuffixBadURL(t *testing.T) {
	req, _ := http.NewRequest("GET", "/search?suffx=-3", nil)
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)
	checkResponseCode(t, http.StatusBadRequest, response.Code)
}
