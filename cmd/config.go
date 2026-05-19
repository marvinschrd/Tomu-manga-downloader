package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mangatool/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Printf("  %-20s %s\n", dim("output_dir"), cfg.OutputDir)
		fmt.Printf("  %-20s %s\n", dim("default_language"), cfg.DefaultLanguage)
		fmt.Printf("  %-20s %s\n", dim("default_source"), cfg.DefaultSource)
		fmt.Println()
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}
		switch args[0] {
		case "output_dir":
			cfg.OutputDir = args[1]
		case "default_language":
			cfg.DefaultLanguage = args[1]
		case "default_source":
			cfg.DefaultSource = args[1]
		default:
			return fmt.Errorf("unknown config key %q — valid keys: output_dir, default_language, default_source", args[0])
		}
		if err := config.Save(cfg); err != nil {
			return err
		}
		fmt.Printf("  %s  %s = %s\n", green("✅"), bold(args[0]), args[1])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
}
