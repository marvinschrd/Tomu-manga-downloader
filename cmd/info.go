package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"mangatool/core"
	"mangatool/core/sources/mangadex"
)

var infoCmd = &cobra.Command{
	Use:   "info <manga-id>",
	Short: "Show manga details and chapter list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := mangadex.New()
		manga, err := source.GetManga(args[0])
		if err != nil {
			return err
		}

		fmt.Printf("Title:  %s\n", manga.Title)
		fmt.Printf("Status: %s\n", manga.Status)
		fmt.Printf("ID:     %s\n", manga.ID)
		if manga.Description != "" {
			fmt.Printf("\n%s\n", manga.Description)
		}

		lang, _ := cmd.Flags().GetString("lang")
		chapters, err := source.GetChapters(manga.ID, core.ChapterOptions{Language: lang})
		if err != nil {
			return err
		}
		fmt.Printf("\nChapters (%d):\n", len(chapters))
		for _, ch := range chapters {
			vol := ""
			if ch.Volume != "" {
				vol = fmt.Sprintf(" [Vol.%s]", ch.Volume)
			}
			title := ""
			if ch.Title != "" {
				title = " — " + ch.Title
			}
			fmt.Printf("  Ch.%-6v%s%s\n", ch.Number, vol, title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP("lang", "l", "", "language filter (default: all languages)")
}
