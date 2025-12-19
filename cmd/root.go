/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "labman",
	Short: "Homelab server management CLI",
	Long: `labman connects to your homelab hosts over SSH and runs curated workflows
such as logging in, checking cluster health, and inspecting Kubernetes state.
Use it as the entry point for every cluster subcommand.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.labman/config.yaml or $XDG_CONFIG_HOME/labman/config.yaml)")
}
