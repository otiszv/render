package main

import (
	"gitlab.uaus.cn/devops/jenkinsrender/cmd"
)

var (
	version   string
	buildDate string
)

func main() {
	cmd.Execute(version, buildDate)
}
