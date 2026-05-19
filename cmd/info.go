package cmd

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
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

		spinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("  Fetching manga info..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetWriter(os.Stderr),
		)
		spinner.Add(1)

		manga, err := source.GetManga(args[0])
		if err != nil {
			spinner.Clear()
			return err
		}

		lang, _ := cmd.Flags().GetString("lang")
		chapters, err := source.GetChapters(manga.ID, core.ChapterOptions{Language: lang})
		spinner.Clear()
		if err != nil {
			return err
		}

		fmt.Printf("\n  %s  %s\n", "📖", bold(manga.Title))
		fmt.Printf("  %s  %s\n", dim("Status"), manga.Status)
		fmt.Printf("  %s  %s\n", dim("ID    "), dim(manga.ID))
		if manga.Description != "" {
			fmt.Printf("\n  %s\n", dim(manga.Description))
		}

		fmt.Printf("\n  %s\n\n", bold(fmt.Sprintf("Chapters (%d):", len(chapters))))
		for _, ch := range chapters {
			vol := ""
			if ch.Volume != "" {
				vol = dim(fmt.Sprintf(" [Vol.%s]", ch.Volume))
			}
			title := ""
			if ch.Title != "" {
				title = dim(" — " + ch.Title)
			}
			fmt.Printf("  Ch.%-6v%s%s\n", ch.Number, vol, title)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().StringP("lang", "l", "", "language filter (default: all languages)")
}
