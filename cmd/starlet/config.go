package main

import (
	"fmt"
	"strings"

	cl "bitbucket.org/ai69/colorlogo"
	"github.com/1set/gut/yos"
	"github.com/1set/gut/ystring"
)

var (
	CIBuildNum string
	BuildDate  string
	BuildHost  string
	GoVersion  string
	GitBranch  string
	GitCommit  string
	GitSummary string
)

var (
	logoArt = `
  ███████╗████████╗ █████╗ ██████╗ ██╗     ███████╗████████╗
  ██╔════╝╚══██╔══╝██╔══██╗██╔══██╗██║     ██╔════╝╚══██╔══╝
  ███████╗   ██║   ███████║██████╔╝██║     █████╗     ██║
  ╚════██║   ██║   ██╔══██║██╔══██╗██║     ██╔══╝     ██║
  ███████║   ██║   ██║  ██║██║  ██║███████╗███████╗   ██║
  ╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝   ╚═╝
`
	logoArtColor = cl.CherryBlossomsByColumn(logoArt)
)

func displayBuildInfo() {
	// write logo
	var sb strings.Builder
	sb.WriteString(logoArtColor)
	sb.WriteString(ystring.NewLine)

	// inline helpers
	arrow := "➣ "
	if yos.IsOnWindows() {
		arrow = "> "
	}
	addNonBlankField := func(name, value string) {
		if ystring.IsNotBlank(value) {
			fmt.Fprintln(&sb, arrow+name+":", value)
		}
	}

	addNonBlankField("Build Num ", CIBuildNum)
	addNonBlankField("Build Date", BuildDate)
	addNonBlankField("Build Host", BuildHost)
	addNonBlankField("Go Version", GoVersion)
	addNonBlankField("Git Branch", GitBranch)
	addNonBlankField("Git Commit", GitCommit)
	addNonBlankField("GitSummary", GitSummary)

	fmt.Println(sb.String())
}
