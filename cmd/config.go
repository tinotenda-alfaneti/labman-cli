package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tinotenda-alfaneti/labman/internal/config"
	"gopkg.in/yaml.v2"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage labman configuration",
	Long: `View, validate, and initialize the labman configuration file.
The config file is located at ~/.labman/config.yaml (or $XDG_CONFIG_HOME/labman/config.yaml).`,
	Run: func(cmd *cobra.Command, args []string) {
		printBanner(cmd)
		printSection(cmd, "CONFIG", "Use 'labman config show', 'labman config validate', or 'labman config init'")
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display the current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		configPath, _ := config.GetConfigPath()

		// Marshal to YAML for display
		yamlData, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}

		output := fmt.Sprintf("Configuration file: %s\n\n%s", configPath, string(yamlData))
		printSection(cmd, "CONFIGURATION", output)

		return nil
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the configuration file for errors",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)

		configPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("get config path: %w", err)
		}

		// Check if config file exists
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			printSection(cmd, "VALIDATION", fmt.Sprintf("No config file found at %s\nRun 'labman config init' to create one.", configPath))
			return nil
		}

		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			printSection(cmd, "VALIDATION FAILED", fmt.Sprintf("❌ %v", err))
			return fmt.Errorf("config validation failed")
		}

		summary := fmt.Sprintf("✓ Configuration is valid\n✓ %d hosts configured\n✓ %d groups defined",
			len(cfg.Hosts), len(cfg.Groups))
		printSection(cmd, "VALIDATION SUCCESS", summary)

		return nil
	},
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a sample configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		printBanner(cmd)

		if err := config.CreateDefaultConfig(); err != nil {
			return fmt.Errorf("create config: %w", err)
		}

		configPath, _ := config.GetConfigPath()
		printSection(cmd, "CONFIG INITIALIZED", fmt.Sprintf("Created sample configuration at:\n%s\n\nEdit this file to add your hosts and preferences.", configPath))

		return nil
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show the configuration file path",
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath, err := config.GetConfigPath()
		if err != nil {
			return fmt.Errorf("get config path: %w", err)
		}

		exists := "does not exist"
		if _, err := os.Stat(configPath); err == nil {
			exists = "exists"
		}

		printBanner(cmd)
		printSection(cmd, "CONFIG PATH", fmt.Sprintf("%s\nStatus: %s", configPath, exists))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configPathCmd)
}
