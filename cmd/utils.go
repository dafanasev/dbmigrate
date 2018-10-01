package main

import (
	"fmt"
	"os"
)

// exitWithError prints an error to the terminal and terminates app with error
func exitWithError(err error) {
	if migrator != nil {
		migrator.Close()
	}

	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func pluralize(s string, n int) string {
	if n != 1 {
		return s + "s"
	}
	return s
}
