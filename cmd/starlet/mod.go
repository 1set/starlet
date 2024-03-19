package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/1set/starlet"
	"github.com/1set/starlet/dataconv"
	shttp "github.com/1set/starlet/lib/http"
	"go.starlark.net/starlark"
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

func setMachineExtras(m *starlet.Machine, args []string) {
	sysLoader := loadSysModule(args)
	m.AddPreloadModules(starlet.ModuleLoaderList{sysLoader})
	m.AddLazyloadModules(starlet.ModuleLoaderMap{"sys": sysLoader})
}

func loadSysModule(args []string) func() (starlark.StringDict, error) {
	// get sa
	sa := make([]starlark.Value, 0, len(args))
	for _, arg := range args {
		sa = append(sa, starlark.String(arg))
	}
	// build module
	sd := starlark.StringDict{
		"platform": starlark.String(runtime.GOOS),
		"arch":     starlark.String(runtime.GOARCH),
		"version":  starlark.MakeUint(starlark.CompilerVersion),
		"argv":     starlark.NewList(sa),
	}
	return dataconv.WrapModuleData("sys", sd)
}
