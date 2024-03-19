// A simple example of using the Starlet REPL or running a script.
package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"bitbucket.org/neiku/winornot"
	"github.com/1set/gut/ystring"
	"github.com/1set/starlet"
	shttp "github.com/1set/starlet/lib/http"
	flag "github.com/spf13/pflag"
	"go.starlark.net/starlark"
	"golang.org/x/term"
)

var (
	allowRecursion      bool
	allowGlobalReassign bool
	preloadModules      []string
	lazyLoadModules     []string
	includePath         string
	codeContent         string
	webPort             uint16
)

var (
	defaultPreloadModules = starlet.GetAllBuiltinModuleNames()
)

func init() {
	flag.BoolVarP(&allowRecursion, "recursion", "r", false, "allow recursion in Starlark code")
	flag.BoolVarP(&allowGlobalReassign, "globalreassign", "g", false, "allow reassigning global variables in Starlark code")
	flag.StringSliceVarP(&preloadModules, "preload", "p", defaultPreloadModules, "preload modules before executing Starlark code")
	flag.StringSliceVarP(&lazyLoadModules, "lazyload", "l", defaultPreloadModules, "lazy load modules when executing Starlark code")
	flag.StringVarP(&includePath, "include", "i", ".", "include path for Starlark code to load modules from")
	flag.StringVarP(&codeContent, "code", "c", "", "Starlark code to execute")
	flag.Uint16VarP(&webPort, "web", "w", 0, "run web server on specified port, it provides reader,writer,fprintf functions for Starlark code to use")
	flag.Parse()

	// fix for Windows terminal output
	winornot.EnableANSIControl()
}

func main() {
	os.Exit(processArgs())
}

func processArgs() int {
	// get starlet machine
	mac := starlet.NewWithNames(nil, preloadModules, lazyLoadModules)
	if allowRecursion {
		mac.EnableRecursionSupport()
	}
	if allowGlobalReassign {
		mac.EnableGlobalReassign()
	}

	// for local modules
	var incFS fs.FS
	if ystring.IsNotBlank(includePath) {
		incFS = os.DirFS(includePath)
	}

	// check arguments
	nargs := flag.NArg()
	hasCode := ystring.IsNotBlank(codeContent)
	switch {
	case webPort > 0:
		// run web server
		var setCode func(m *starlet.Machine)
		if hasCode {
			// run code string from argument
			setCode = func(m *starlet.Machine) {
				m.SetScript("web.star", []byte(codeContent), incFS)
			}
		} else if nargs == 1 {
			// run code from file
			fileName := flag.Arg(0)
			setCode = func(m *starlet.Machine) {
				m.SetScript(fileName, nil, incFS)
			}
		} else {
			// no code to run
			PrintError(fmt.Errorf("no code to run as web server"))
			return 1
		}
		// start web server
		if err := runWebServer(webPort, setCode); err != nil {
			PrintError(err)
			return 1
		}
	case nargs == 0 && hasCode:
		// run code string from argument
		mac.SetScript("direct.star", []byte(codeContent), incFS)
		_, err := mac.Run()
		if err != nil {
			PrintError(err)
			return 1
		}
	case nargs == 0 && !hasCode:
		// run REPL
		stdinIsTerminal := term.IsTerminal(int(os.Stdin.Fd()))
		if stdinIsTerminal {
			displayBuildInfo()
		}
		mac.SetScript("repl", nil, incFS)
		mac.REPL()
		if stdinIsTerminal {
			fmt.Println()
		}
	case nargs == 1:
		// run code from file
		fileName := flag.Arg(0)
		mac.SetScript(fileName, nil, incFS)
		_, err := mac.Run()
		if err != nil {
			PrintError(err)
			return 1
		}
	case nargs > 1:
		fmt.Println(`want at most one Starlark file name`)
		return 1
	default:
		flag.Usage()
		return 1
	}
	return 0
}

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

// PrintError prints the error to stderr,
// or its backtrace if it is a Starlark evaluation error.
func PrintError(err error) {
	if evalErr, ok := err.(*starlark.EvalError); ok {
		fmt.Fprintln(os.Stderr, evalErr.Backtrace())
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
