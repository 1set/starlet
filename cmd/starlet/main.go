package main

import (
	"fmt"
	"os"

	"bitbucket.org/neiku/winornot"
	"github.com/1set/gut/ystring"
	flag "github.com/spf13/pflag"
)

var (
	allowRecursion      bool
	allowGlobalReassign bool
	preloadModules      []string
	lazyLoadModules     []string
	includePath         string
	codeContent         string
)

func init() {
	flag.BoolVarP(&allowRecursion, "recursion", "r", false, "allow recursion in Starlark code")
	flag.BoolVarP(&allowGlobalReassign, "globalreassign", "g", false, "allow reassigning global variables in Starlark code")
	flag.StringSliceVarP(&preloadModules, "preload", "p", nil, "preload modules before executing Starlark code")
	flag.StringSliceVarP(&lazyLoadModules, "lazyload", "l", nil, "lazy load modules when executing Starlark code")
	flag.StringVarP(&includePath, "include", "i", ".", "include path for Starlark code to load modules from")
	flag.StringVarP(&codeContent, "code", "c", "", "Starlark code to execute")
	flag.Parse()

	// fix for Windows terminal output
	winornot.EnableANSIControl()
	displayBuildInfo() // TODO: move inside
}

func main() {
	os.Exit(processArgs())
}

func processArgs() int {
	nargs := flag.NArg()
	hasCode := ystring.IsNotBlank(codeContent)
	switch {
	case nargs == 0 && hasCode:
	// run code from command line
	case nargs == 0 && !hasCode:
		// run REPL
	case nargs == 1:
	// run code from file
	case nargs > 1:
		fmt.Println(`want at most one Starlark file name`)
	default:
		flag.Usage()
	}

	return 0
}
