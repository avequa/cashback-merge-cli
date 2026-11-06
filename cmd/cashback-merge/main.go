// Package main implements the Cashback Merge CLI application.
package main

import (
	"cashbackmerge/internal/cli"
	"os"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
