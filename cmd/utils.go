package main

import (
	"fmt"
	"os"
)

// exitWithError prints an error to the terminal and terminates app with error
func exitWithError(err error) {
	migrator.Close()

	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
