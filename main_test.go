package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Router() *mux.Router {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/search", search)
	myRouter.HandleFunc("/getall", getall)
	myRouter.HandleFunc("/set", createNewEntity).Methods("POST")
	myRouter.HandleFunc("/get/{key}", returnSingleEntityValue)
	// log.Fatal(http.ListenAndServe(":10000", myRouter))
	return myRouter
}

//quering for key "abc-1" which is not present.
func TestGetEntityBad(t *testing.T) {
	request, _ := http.NewRequest("GET", "/get/abc-1", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 404, response.Code, "Expected 404 status code")
}

func TestGetEntityGood(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response1 := httptest.NewRecorder()
	Router().ServeHTTP(response1, request)
	request, _ = http.NewRequest("GET", "/get/abc-1", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Expected 200 status code")
}

// POST request with invalid json body
func TestSetEntityBad(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1","hello": "test"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "Expected 400 status code")
}

func TestSetEntityGood(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Expected 200 status code")
}

// searching with prefix
func TestSearchPrefixGood(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	request, _ = http.NewRequest("GET", "/search?prefix=abc", nil)
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Expected 200 status code")
}

func TestSearchPrefixBad(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response1 := httptest.NewRecorder()
	Router().ServeHTTP(response1, request)
	request, _ = http.NewRequest("GET", "/search?prefix=xyz", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 404, response.Code, "Expected 404 status code")
}

func TestSearchPrefixBadURL(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response1 := httptest.NewRecorder()
	Router().ServeHTTP(response1, request)
	request, _ = http.NewRequest("GET", "/search?prefi", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "Expected 400 status code")
}

// searching with suffix
func TestSearchSuffixGood(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response1 := httptest.NewRecorder()
	Router().ServeHTTP(response1, request)
	request, _ = http.NewRequest("GET", "/search?suffix=-1", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Expected 200 status code")
}

func TestSearchSuffixBad(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response1 := httptest.NewRecorder()
	Router().ServeHTTP(response1, request)
	request, _ = http.NewRequest("GET", "/search?suffix=-2", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 404, response.Code, "Expected 404 status code")
}

func TestSearchSuffixBadURL(t *testing.T) {
	var jsonStr = []byte(`{"key":"abc-1", "value":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response1 := httptest.NewRecorder()
	Router().ServeHTTP(response1, request)
	request, _ = http.NewRequest("GET", "/search?suffixx=-2", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 400, response.Code, "Expected 400 status code")
}
