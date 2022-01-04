// main.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Entity - struct for all entities
// an entity has a key-string and a value-string
type Entity struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

var Entities []Entity

func getall(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint Hit: getall")
	json.NewEncoder(w).Encode(Entities)
}

func returnSingleEntityValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]
	var keyFound bool
	for _, entity := range Entities {
		if entity.Key == key {
			keyFound = true
			json.NewEncoder(w).Encode(entity.Value)
		}
	}
	if keyFound {
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("Get request: invalid key")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - No data found!"))
	}
}

func createNewEntity(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// unmarshal this into a new Entity struct
	// append this to our Entities array.
	reqBody, _ := ioutil.ReadAll(r.Body)
	var entity Entity
	json.Unmarshal(reqBody, &entity)
	if entity.Key == "" || entity.Value == "" {
		log.Println("Post request: Invalid request body")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad request!"))
	} else {
		Entities = append(Entities, entity)
		json.NewEncoder(w).Encode(entity)
		w.WriteHeader(http.StatusAccepted)
	}

	// update our global Entities array to include
	// our new Entity

}

func search(w http.ResponseWriter, r *http.Request) {
	RawQuery := r.URL.RawQuery
	var keyFound bool
	if strings.HasPrefix(RawQuery, "prefix=") {
		keys, ok := r.URL.Query()["prefix"]
		// search?prefix=abc
		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'prefix' is missing")
			return
		}

		key := keys[0]
		log.Println("Url Param 'prefix' is: " + string(key))

		for _, entity := range Entities {
			if strings.HasPrefix(entity.Key, key) {
				json.NewEncoder(w).Encode(entity)
				keyFound = true
			}
		}

	} else if strings.HasPrefix(RawQuery, "suffix=") {
		keys, ok := r.URL.Query()["suffix"]
		// /search?suffix=-1
		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'suffix' is missing")
			return
		}

		key := keys[0]
		log.Println("Url Param 'suffix' is: " + string(key))

		for _, entity := range Entities {
			if strings.HasSuffix(entity.Key, key) {
				json.NewEncoder(w).Encode(entity)
				keyFound = true
			}
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad Request!"))
		log.Println("Invalid search parameters.")
		return
	}
	if keyFound {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - No data found!"))
	}
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/search", search)
	myRouter.HandleFunc("/getall", getall)
	myRouter.HandleFunc("/set", createNewEntity).Methods("POST")
	myRouter.HandleFunc("/get/{key}", returnSingleEntityValue)
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func main() {
	// dummy data
	Entities = []Entity{
		Entity{Key: "abc-1", Value: "Thermodynamics"},
		Entity{Key: "abc-2", Value: "Automotive Engineering"},
	}
	handleRequests()
}
