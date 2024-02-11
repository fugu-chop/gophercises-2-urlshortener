package main

import (
	"log"
	"net/http"

	"gopkg.in/yaml.v2"
)

type ParsedYaml struct {
	Path string `yaml:"path"`
	URL  string `yaml:"url"`
}

// MapHandler will return an http.HandlerFunc (which also
// implements http.Handler) that will attempt to map any
// paths (keys in the map) to their corresponding URL (values
// that each key in the map points to, in string format).
// If the path is not provided in the map, then the fallback
// http.Handler will be called instead.
func MapHandler(pathsToUrls map[string]string, fallback http.Handler) http.HandlerFunc {
	// This is the boilerplate wrapping code that returns a
	// http.HandlerFunc function type
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		redirectUrl, ok := pathsToUrls[r.URL.Path]
		if ok {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	})
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
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse YAML
		// We require a slice return type as otherwise only the
		// first entry in the YAML is retained
		var output []ParsedYaml

		err := yaml.Unmarshal(yml, &output)
		if err != nil {
			log.Fatalf("cannot unmarshal yaml data: %v", err)
		}

		// check if path exists
		pathMap := parseYamlToMap(output)
		// Attempting to use MapHandler here doesn't appear
		// to work - there is some weird double redirect
		// happening under the hood somewhere
		redirectUrl, ok := pathMap[r.URL.Path]
		if ok {
			http.Redirect(w, r, redirectUrl, http.StatusFound)
		} else {
			fallback.ServeHTTP(w, r)
		}
	}), nil
}

func parseYamlToMap(parsedYaml []ParsedYaml) map[string]string {
	var shortenerKeys = make(map[string]string)

	for _, entry := range parsedYaml {
		shortenerKeys[entry.Path] = entry.URL
	}

	return shortenerKeys
}
