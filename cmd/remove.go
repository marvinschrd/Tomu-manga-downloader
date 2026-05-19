package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"mangatool/core/library"
)

var removeCmd = &cobra.Command{
	Use:   "remove <manga-id>",
	Short: "Remove manga from library and delete files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open(libraryPath())
		if err != nil {
			return err
		}
		defer lib.Close()

		m, err := lib.GetManga(args[0])
		if err != nil {
			return err
		}

		dbOnly, _ := cmd.Flags().GetBool("db-only")

		fmt.Println()
		if dbOnly {
			fmt.Printf("  %s  Remove %s from library? Files on disk will be kept. %s ",
				yellow("⚠"), bold(m.Title), dim("[y/N]"))
		} else {
			fmt.Printf("  %s  Delete %s and all files in %s? %s ",
				yellow("⚠"), bold(m.Title), dim(m.OutputDir), dim("[y/N]"))
		}

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Printf("  %s  Cancelled.\n\n", dim("↩"))
			return nil
		}

		if !dbOnly {
			if err := os.RemoveAll(m.OutputDir); err != nil {
				return fmt.Errorf("deleting files: %w", err)
			}
			fmt.Printf("  %s  Deleted %s\n", "🗑", dim(m.OutputDir))
		}

		if err := lib.RemoveManga(m.ID); err != nil {
			return fmt.Errorf("removing from library: %w", err)
		}
		fmt.Printf("  %s  %s removed from library.\n\n", green("✅"), bold(m.Title))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().Bool("db-only", false, "remove from library without deleting files on disk")
}
