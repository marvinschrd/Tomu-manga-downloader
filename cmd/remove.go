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

		if dbOnly {
			fmt.Printf("Remove %q from library (files on disk kept)? [y/N] ", m.Title)
		} else {
			fmt.Printf("Delete %q and all files in %s? [y/N] ", m.Title, m.OutputDir)
		}

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("Cancelled.")
			return nil
		}

		if !dbOnly {
			if err := os.RemoveAll(m.OutputDir); err != nil {
				return fmt.Errorf("deleting files: %w", err)
			}
			fmt.Printf("Deleted %s\n", m.OutputDir)
		}

		if err := lib.RemoveManga(m.ID); err != nil {
			return fmt.Errorf("removing from library: %w", err)
		}
		fmt.Printf("Removed %q from library.\n", m.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
	removeCmd.Flags().Bool("db-only", false, "remove from library without deleting files on disk")
}
