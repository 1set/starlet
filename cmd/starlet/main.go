// A simple example of using the Starlet REPL or running a script.
package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"bitbucket.org/neiku/winornot"
	"github.com/1set/gut/ystring"
	"github.com/1set/starlet"
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
	argCode := ystring.IsNotBlank(codeContent)
	switch {
	case webPort > 0:
		// run web server
		var setCode func(m *starlet.Machine)
		if argCode {
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
	case argCode:
		// run code string from argument
		setMachineExtras(mac, append([]string{`-c`}, flag.Args()...))
		mac.SetScript("direct.star", []byte(codeContent), incFS)
		_, err := mac.Run()
		if err != nil {
			PrintError(err)
			return 1
		}
	case nargs == 0 && !argCode:
		// run REPL
		stdinIsTerminal := term.IsTerminal(int(os.Stdin.Fd()))
		if stdinIsTerminal {
			displayBuildInfo()
		}
		setMachineExtras(mac, []string{``})
		mac.SetScript("repl", nil, incFS)
		mac.REPL()
		if stdinIsTerminal {
			fmt.Println()
		}
	case nargs >= 1:
		// run code from file
		fileName := flag.Arg(0)
		bs, err := ioutil.ReadFile(fileName)
		if err != nil {
			PrintError(err)
			return 1
		}
		setMachineExtras(mac, flag.Args())
		mac.SetScript(filepath.Base(fileName), bs, incFS)
		if _, err := mac.Run(); err != nil {
			PrintError(err)
			return 1
		}
	default:
		flag.Usage()
		return 1
	}
	return 0
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
