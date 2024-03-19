package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/1set/starlet"
	shttp "github.com/1set/starlet/lib/http"
)

func runWebServer(port uint16, setCode func(m *starlet.Machine)) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// prepare envs
		resp := shttp.NewServerResponse()
		glb := starlet.StringAnyMap{
			"request":  shttp.ConvertServerRequest(r),
			"response": resp.Struct(),
		}

		// run code
		mac := starlet.NewWithNames(glb, preloadModules, lazyLoadModules)
		setCode(mac)
		_, err := mac.Run()

		// handle error
		if err != nil {
			log.Printf("Runtime Error: %v\n", err)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprintf(w, "Runtime Error: %v", err); err != nil {
				log.Printf("Error writing response: %v", err)
			}
			return
		}

		// handle response
		if err = resp.Write(w); err != nil {
			w.Header().Add("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}
	})

	log.Printf("Server is starting on port: %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
	return err
}

func runWebServerLegacy(port uint16, setCode func(m *starlet.Machine)) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		glb := starlet.StringAnyMap{
			"reader":  r,
			"writer":  w,
			"fprintf": fmt.Fprintf,
		}

		mac := starlet.NewWithNames(glb, preloadModules, lazyLoadModules)
		setCode(mac)
		//mac.SetScript("web.star", code, incFS)
		if _, err := mac.Run(); err != nil {
			log.Printf("Runtime Error: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprintf(w, "Runtime Error: %v", err); err != nil {
				log.Printf("Error writing response: %v", err)
			}
			return
		}
	})

	log.Printf("Server is starting on port: %d\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
	return err
}
