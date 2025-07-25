package server

import "github.com/alecthomas/kingpin"

var (
	defaultSeverVersion  = "1.0.0"
	enableDefaultPromReg *bool
)

func init() {
	enableDefaultPromReg = kingpin.Flag(
		"enable-default-prom-reg",
		"enable default prom reg").
		Bool()
}
