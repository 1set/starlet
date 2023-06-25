// A simple example of using the Starlet REPL or running a script.
package main

import (
	"fmt"
	"io/fs"
	"os"

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
)

var (
	defaultPreloadModules = starlet.GetAllBuiltinModuleNames()
)

func init() {
	flag.BoolVarP(&allowRecursion, "recursion", "r", false, "allow recursion in Starlark code")
	flag.BoolVarP(&allowGlobalReassign, "globalreassign", "g", false, "allow reassigning global variables in Starlark code")
	flag.StringSliceVarP(&preloadModules, "preload", "p", defaultPreloadModules, "preload modules before executing Starlark code")
	flag.StringSliceVarP(&lazyLoadModules, "lazyload", "l", nil, "lazy load modules when executing Starlark code")
	flag.StringVarP(&includePath, "include", "i", ".", "include path for Starlark code to load modules from")
	flag.StringVarP(&codeContent, "code", "c", "", "Starlark code to execute")
	flag.Parse()

	// fix for Windows terminal output
	winornot.EnableANSIControl()
}

func main() {
	os.Exit(processArgs())
}

func processArgs() int {
	// get starlet machine
	if allowRecursion {
		starlet.EnableRecursionSupport()
	}
	if allowGlobalReassign {
		starlet.EnableGlobalReassign()
	}
	mac := starlet.NewWithNames(nil, preloadModules, lazyLoadModules)
	var incFS fs.FS
	if ystring.IsNotBlank(includePath) {
		incFS = os.DirFS(includePath)
	}

	// check arguments
	nargs := flag.NArg()
	hasCode := ystring.IsNotBlank(codeContent)
	switch {
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

// PrintError prints the error to stderr,
// or its backtrace if it is a Starlark evaluation error.
func PrintError(err error) {
	if evalErr, ok := err.(*starlark.EvalError); ok {
		fmt.Fprintln(os.Stderr, evalErr.Backtrace())
	} else {
		fmt.Fprintln(os.Stderr, err)
	}
}
