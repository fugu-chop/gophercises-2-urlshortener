package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

func main() {
	var handler http.HandlerFunc

	// Define flags
	yamlPtr := flag.String("yamlImport", "", "where the handler should look to import yaml routes")
	jsonPtr := flag.String("jsonImport", "", "where the handler should look to import json routes")
	dbPtr := flag.Bool("dbImport", false, "whether or not the handler should access a BoltDB instance for routes")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := MapHandler(pathsToUrls, mux)

	// import Yaml file based on flag presence
	if *yamlPtr != "" {
		yamlFile, err := os.ReadFile(*yamlPtr)
		if err != nil {
			log.Fatalf("cannot open yaml file: %v", err)
		}
		handler, err = YAMLHandler(yamlFile, mapHandler)
		if err != nil {
			panic(err)
		}
		// import JSON file based on flag presence
	} else if *jsonPtr != "" {
		jsonFile, err := os.ReadFile(*jsonPtr)
		if err != nil {
			log.Fatalf("cannot open yaml file: %v", err)
		}
		handler, err = JSONHandler(jsonFile, mapHandler)
		if err != nil {
			panic(err)
		}
		// Use BoltDB if flag is enabled
	} else if *dbPtr {
		db, err := bolt.Open(
			"routes.db",
			0600,
			&bolt.Options{Timeout: 2 * time.Second},
		)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		seedDB(db, pathsToUrls)
	} else {
		handler = mapHandler
	}

	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", handler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
