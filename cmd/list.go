package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mangatool/core/library"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all downloaded manga",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.Open(libraryPath())
		if err != nil {
			return err
		}
		defer lib.Close()

		manga, err := lib.ListManga()
		if err != nil {
			return err
		}
		if len(manga) == 0 {
			fmt.Println("No manga in library. Use 'download' to add some.")
			return nil
		}
		for _, m := range manga {
			chapters, _ := lib.GetChapters(m.ID)
			fmt.Printf("[%s] %s — %d chapters in %s\n", m.ID, m.Title, len(chapters), m.OutputDir)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
