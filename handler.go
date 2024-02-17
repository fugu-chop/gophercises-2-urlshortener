package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"gopkg.in/yaml.v2"
)

const bucketName = "Routes"

type ParsedFile struct {
	Path string `yaml:"path" json:"path"`
	URL  string `yaml:"url" json:"url"`
}

func DBHandler(db *bolt.DB, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mappedRoute string
		err := db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(bucketName))
			dbResult := b.Get([]byte(r.URL.Path))
			if dbResult != nil {
				mappedRoute = string(dbResult)
			}

			return nil
		})
		if err != nil {
			log.Fatalf("cannot query database: %v", err)
		}

		if mappedRoute == "" {
			fallback.ServeHTTP(w, r)
		} else {
			http.Redirect(w, r, mappedRoute, http.StatusFound)
		}
	}
}

func JSONHandler(jsonInput []byte, fallback http.Handler) (http.HandlerFunc, error) {
	var parsedJson []ParsedFile
	err := json.Unmarshal(jsonInput, &parsedJson)
	if err != nil {
		log.Fatalf("cannot unmarshal json data: %v", err)
	}

	parsedMap := parseFileToMap(parsedJson)
	return MapHandler(parsedMap, fallback), nil
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		redirectUrl, ok := pathsToUrls[r.URL.Path]
		if ok {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}
}

// YAMLHandler will parse the provided YAML and then return
// an http.HandlerFunc (which also implements http.Handler)
// that will attempt to map any paths to their corresponding
// URL. If the path is not provided in the YAML, then the
// fallback http.Handler will be called instead.
//
// YAML is expected to be in the format:
//
//   - path: /some-path
//     url: https://www.some-url.com/demo
//
// The only errors that can be returned all related to having
// invalid YAML data.
//
// See MapHandler to create a similar http.HandlerFunc via
// a mapping of paths to urls.
func YAMLHandler(yml []byte, fallback http.Handler) (http.HandlerFunc, error) {
	// Parse YAML
	// We require a slice return type as otherwise only the
	// first entry in the YAML is retained
	var output []ParsedFile

	err := yaml.Unmarshal(yml, &output)
	if err != nil {
		log.Fatalf("cannot unmarshal yaml data: %v", err)
	}

	// check if path exists
	pathMap := parseFileToMap(output)
	return MapHandler(pathMap, fallback), nil
}

func parseFileToMap(parsedFile []ParsedFile) map[string]string {
	var shortenerKeys = make(map[string]string)

	for _, entry := range parsedFile {
		shortenerKeys[entry.Path] = entry.URL
	}

	return shortenerKeys
}

func seedDB(db *bolt.DB, routes map[string]string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Fatalf("cannot create bucket in database: %v", err)
		}

		for k, v := range routes {
			b.Put([]byte(k), []byte(v))
		}

		return nil
	})

	if err != nil {
		log.Fatalf("cannot seed database: %v", err)
	}

	return nil
}
