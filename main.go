// main.go
package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Entity - struct for all entities
// an entity has a key-string and a value-string
type Entity struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

var mutex = &sync.RWMutex{}

type App struct {
	Router   *mux.Router
	Entities map[string]string
}

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

func (a *App) returnSingleEntityValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyName := vars["key"]
	if value, ok := a.Entities[keyName]; ok {
		respondWithJSON(w, http.StatusOK, value, "get")
	} else {
		w.WriteHeader(http.StatusNotFound)
		respondWithError(w, http.StatusNotFound, "Data not found", "get")
	}
}

func (a *App) createNewEntity(w http.ResponseWriter, r *http.Request) {
	reqBody, _ := ioutil.ReadAll(r.Body)
	var entity Entity
	json.Unmarshal(reqBody, &entity)
	if entity.Key == "" || entity.Value == "" {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Post request: Invalid request body", "set")
	} else {
		w.WriteHeader(http.StatusCreated)
		go a.writeToMap(entity, w)
		time.Sleep(1 * time.Millisecond)
	}

}

func (a *App) writeToMap(entity Entity, w http.ResponseWriter) {
	mutex.Lock()
	a.Entities[entity.Key] = entity.Value
	mutex.Unlock()
	respondWithJSON(w, http.StatusCreated, entity, "set")
}

func (a *App) search(w http.ResponseWriter, r *http.Request) {
	RawQuery := r.URL.RawQuery
	var keyFound bool
	var entity Entity
	var entities []Entity

	if strings.HasPrefix(RawQuery, "prefix=") {
		keys, ok := r.URL.Query()["prefix"]
		// search?prefix=<prefix>
		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'prefix' is missing")
			w.WriteHeader(http.StatusBadRequest)
			respondWithError(w, http.StatusBadRequest, "Bad URL request", "/search")
			return
		}

		keyPrefix := keys[0]

		for key, value := range a.Entities {
			if strings.HasPrefix(key, keyPrefix) {
				// json.NewEncoder(w).Encode(key + " = " + value)
				entity.Key = key
				entity.Value = value
				entities = append(entities, entity)
				keyFound = true
			}
		}

	} else if strings.HasPrefix(RawQuery, "suffix=") {
		keys, ok := r.URL.Query()["suffix"]
		// /search?suffix=<suffix>
		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'suffix' is missing")
			w.WriteHeader(http.StatusBadRequest)
			respondWithError(w, http.StatusBadRequest, "Bad URL request", "search")
			return
		}

		keySuffix := keys[0]

		for key, value := range a.Entities {
			if strings.HasSuffix(key, keySuffix) {
				// json.NewEncoder(w).Encode(key + " = " + value)
				entity.Key = key
				entity.Value = value
				entities = append(entities, entity)
				keyFound = true
			}
		}

	} else {
		w.WriteHeader(http.StatusBadRequest)
		respondWithError(w, http.StatusBadRequest, "Bad URL request", "search")
		return
	}
	if keyFound {
		w.WriteHeader(http.StatusOK)
		respondWithJSON(w, http.StatusOK, entities, "search")
	} else {
		w.WriteHeader(http.StatusNotFound)
		respondWithError(w, http.StatusNotFound, "No data found!", "search")
	}
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

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func respondWithError(w http.ResponseWriter, code int, message, endpoint string) {
	getCounter.WithLabelValues(strconv.Itoa(code)).Inc()
	respondWithJSON(w, code, map[string]string{"error": message}, endpoint)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}, endpoint string) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
	switch endpoint {
	case "get":
		getCounter.WithLabelValues(strconv.Itoa(code)).Inc()
	case "set":
		postCounter.WithLabelValues(strconv.Itoa(code)).Inc()
	case "search":
		searchCounter.WithLabelValues(strconv.Itoa(code)).Inc()
	}
}

func (a *App) InitializeData() {
	a.Entities = make(map[string]string)
	a.Entities["abc-1"] = "Thermodynamics"
	a.Entities["abc-2"] = "Automotive Engineering"
}

func (a *App) Initialize() {
	// dummy data
	a.InitializeData()
	for i := 0; i < len(a.Entities); i++ {
		log.Println("Recording existing data ", i)
		keyConter.Inc()
	}
	// Initialize prometheus metrics
	prometheus.MustRegister(getCounter)
	prometheus.MustRegister(postCounter)
	prometheus.MustRegister(searchCounter)

	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/search", a.search)
	a.Router.HandleFunc("/set", a.createNewEntity).Methods("POST")
	a.Router.HandleFunc("/get/{key}", a.returnSingleEntityValue)
	a.Router.Use(prometheusMiddleware)
	a.Router.Path("/metrics").Handler(promhttp.Handler())
	a.Router.HandleFunc("/healthz", healthz)

}

func (a *App) Run(addr string) {
	srv := &http.Server{
		Handler:      a.Router,
		Addr:         addr,
		WriteTimeout: 2 * time.Second,
		ReadTimeout:  2 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func main() {
	a := App{}
	a.Initialize()
	a.Run(":10000")
}
