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

func TestGetEntityBad(t *testing.T) {
	request, _ := http.NewRequest("GET", "/get/abc-1", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 404, response.Code, "Expected 200 status code")
}

func TestGetEntityGood(t *testing.T) {
	var jsonStr = []byte(`{"abc-1":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	request, _ = http.NewRequest("GET", "/get/abc-1", nil)
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Expected 200 status code")
}

func TestSetEntityBad(t *testing.T) {
	var jsonStr = []byte(`{"abc-1":"Some Random Value","hello": "test"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Expected 200 status code")
}

func TestSetEntityGood(t *testing.T) {
	var jsonStr = []byte(`{"abc-1":"Some Random Value"}`)
	request, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(jsonStr))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 202, response.Code, "Expected 202 status code")
}
