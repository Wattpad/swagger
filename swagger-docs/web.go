package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"strings"
	"text/template"

	"github.com/Wattpad/negroni"
)

var host = flag.String("host", "0.0.0.0", "Host")
var port = flag.String("port", "8080", "Port")
var staticContent = flag.String("staticPath", "../swagger-ui", "Path to folder with Swagger UI")
var serviceURL = flag.String("serviceURL", "http://0.0.0.0:8090", "The base path URI of the API service")

func decodeRequestParams(data []byte) (map[string]interface{}, error) {
	userRequest := make(map[string]interface{})
	err := json.Unmarshal(data, &userRequest)
	if err != nil {
		return nil, err
	}
	return userRequest, nil
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	isJSONRequest := false

	if acceptHeaders, ok := r.Header["Accept"]; ok {
		for _, acceptHeader := range acceptHeaders {
			if strings.Contains(acceptHeader, "json") {
				isJSONRequest = true
				break
			}
		}
	}

	if isJSONRequest {
		resourceList, err := decodeRequestParams([]byte(resourceListingJson))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resourceListingJson, err := json.Marshal(resourceList)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write([]byte(resourceListingJson))
	} else {
		http.Redirect(w, r, "/swagger-ui/", http.StatusFound)
	}
}

func apiDescriptionHandler(w http.ResponseWriter, r *http.Request) {
	apiKey := strings.Trim(r.RequestURI, "/")
	if data, ok := apiDescriptionsJson[apiKey]; ok {
		apiDescriptions, err := decodeRequestParams([]byte(data))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		apiDescriptions["basePath"] = *serviceURL
		data, err := json.Marshal(apiDescriptions)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		t, e := template.New("desc").Parse(string(data))
		if e != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		t.Execute(w, *serviceURL)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	flag.Parse()

	// To serve a directory on disk (/tmp) under an alternate URL
	// path (/tmpfiles/), use StripPrefix to modify the request
	// URL's path before the FileServer sees it:
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir(*staticContent))))

	for apiKey := range apiDescriptionsJson {
		mux.HandleFunc("/"+apiKey+"/", apiDescriptionHandler)
	}

	n := negroni.Classic()
	n.UseHandler(mux)
	n.Run(":" + *port)

}
