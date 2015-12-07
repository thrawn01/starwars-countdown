package main

import (
	"flag"
	"fmt"
	"starwars-countdown/vendor/_nuts/github.com/onsi/ginkgo/config"
)

func BuildVersionCommand() *Command {
	return &Command{
		Name:         "version",
		FlagSet:      flag.NewFlagSet("version", flag.ExitOnError),
		UsageCommand: "ginkgo version",
		Usage: []string{
			"Print Ginkgo's version",
		},
		Command: printVersion,
	}
}

func printVersion([]string, []string) {
	fmt.Printf("Ginkgo Version %s\n", config.VERSION)
}
