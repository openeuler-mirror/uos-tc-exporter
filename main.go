package main

import (
	"fmt"
	"os"

	"gitee.com/openeuler/uos-tc-exporter/version"
)

var (
	Name    = "uos_tc_exporter"
	Version = version.Version
)

func main() {

	err := Run(Name, Version)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
