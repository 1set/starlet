package main

import (
	"fmt"
	"os"
	"time"

	"bitbucket.org/neiku/winornot"
	flag "github.com/spf13/pflag"
)

var (
	number int
)

func init() {
	displayBuildInfo()
	flag.IntVarP(&number, "number", "n", 1, "Example of number variable")
	flag.Parse()

	// fix for Windows terminal output
	winornot.EnableANSIControl()
}

func main() {
	pwd, _ := os.Getwd()
	host, _ := os.Hostname()
	fmt.Printf("🌋: Hello, World!\n⏰: %s\n%s\n%s\n", time.Now().Format("2006-01-02T15:04:05-0700"), pwd, host)
}
