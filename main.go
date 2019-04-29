package main

import (
	"github.com/otiszv/render/cmd"
)

var (
	version   string
	buildDate string
)

func main() {
	cmd.Execute(version, buildDate)
}
