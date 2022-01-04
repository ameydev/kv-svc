// main.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Entity - struct for all entities
// an entity has a key-string and a value-string
type Entity struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

var Entities []Entity
var (
	keyConter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "entity_counter",
		Help: "total no. of keys in the DB",
	})
)

var getCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_get_count", // metric name
		Help: "Number of get_key request.",
	},
	[]string{"status"}, // labels
)

var postCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_post_count", // metric name
		Help: "Number of post_key request.",
	},
	[]string{"status"}, // labels
)

var searchCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_request_search_count", // metric name
		Help: "Number of search_key request.",
	},
	[]string{"status"}, // labels
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path"})
)

func getall(w http.ResponseWriter, r *http.Request) {
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
		getCounter.WithLabelValues("200").Inc()
	} else {
		log.Println("Get request: invalid key")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - No data found!"))
		getCounter.WithLabelValues("404").Inc()
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
		postCounter.WithLabelValues("400").Inc()
	} else {
		Entities = append(Entities, entity)
		json.NewEncoder(w).Encode(entity)
		w.WriteHeader(http.StatusAccepted)
		keyConter.Inc()
		postCounter.WithLabelValues("200").Inc()
	}

}

func search(w http.ResponseWriter, r *http.Request) {
	RawQuery := r.URL.RawQuery
	var keyFound bool
	if strings.HasPrefix(RawQuery, "prefix=") {
		keys, ok := r.URL.Query()["prefix"]
		// search?prefix=<prefix>
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
		// /search?suffix=<suffix>
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
		searchCounter.WithLabelValues("400").Inc()
		return
	}
	if keyFound {
		w.WriteHeader(http.StatusOK)
		searchCounter.WithLabelValues("200").Inc()
	} else {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 - No data found!"))
		searchCounter.WithLabelValues("404").Inc()
	}
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/search", search)
	myRouter.HandleFunc("/getall", getall)
	myRouter.HandleFunc("/set", createNewEntity).Methods("POST")
	myRouter.HandleFunc("/get/{key}", returnSingleEntityValue)
	myRouter.Use(prometheusMiddleware)
	myRouter.Path("/metrics").Handler(promhttp.Handler())
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

// prometheusMiddleware implements mux.MiddlewareFunc.
func prometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		timer := prometheus.NewTimer(httpDuration.WithLabelValues(path))
		next.ServeHTTP(w, r)
		timer.ObserveDuration()
	})
}

func initMetrics() {
	// Count the existing keys (dummy data)
	for l, _ := range Entities {
		log.Println("Recording existing data ", l)
		keyConter.Inc()
	}

	prometheus.MustRegister(getCounter)
	prometheus.MustRegister(postCounter)
	prometheus.MustRegister(searchCounter)
}

func main() {
	// dummy data
	Entities = []Entity{
		Entity{Key: "abc-1", Value: "Thermodynamics"},
		Entity{Key: "abc-2", Value: "Automotive Engineering"},
	}
	initMetrics()
	handleRequests()
}
