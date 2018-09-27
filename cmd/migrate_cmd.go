package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var steps int

func init() {
	migrateCmd.LocalFlags().IntVarP(&steps, "steps", "s", 0, "steps")
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate database schema",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viper.ConfigFileUsed())
	},
}
