package main

import (
	"gitlab.uaus.cn/devops/jenkinsfilext/cmd"
)

var (
	version   string
	buildDate string
)

func main() {
	cmd.Execute(version, buildDate)
}
