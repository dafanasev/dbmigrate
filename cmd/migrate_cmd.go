package main

import (
	"github.com/spf13/cobra"
)

var steps int

func init() {
	migrateCmd.LocalFlags().IntVarP(&steps, "steps", "s", 0, "steps")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database schema",
	Run: func(cmd *cobra.Command, args []string) {

	},
}
