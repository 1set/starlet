package main

import (
	"bitbucket.org/neiku/winornot"
	flag "github.com/spf13/pflag"
)

var (
	number int
)

func init() {
	flag.IntVarP(&number, "number", "n", 1, "Example of number variable")
	flag.Parse()

	// fix for Windows terminal output
	winornot.EnableANSIControl()
}

func main() {
	displayBuildInfo()
}
