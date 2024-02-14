package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Define flags
	yamlPtr := flag.String("yamlImport", "", "where the handler should look to import yaml routes")
	flag.Parse()

	mux := defaultMux()

	// Build the MapHandler using the mux as the fallback
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := MapHandler(pathsToUrls, mux)

	// Build the YAMLHandler using the mapHandler as the
	// fallback
	yamlString := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
  `

	fileToConvert := []byte{}

	// import Yaml file based on flag presence
	if *yamlPtr != "" {
		yamlFile, err := os.ReadFile(*yamlPtr)
		if err != nil {
			log.Fatalf("cannot open yaml file: %v", err)
		}
		fileToConvert = yamlFile
	} else {
		fileToConvert = []byte(yamlString)
	}

	yamlHandler, err := YAMLHandler(fileToConvert, mapHandler)
	if err != nil {
		panic(err)
	}
	fmt.Println("Starting the server on :8080")
	http.ListenAndServe(":8080", yamlHandler)
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", hello)
	return mux
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, world!")
}
