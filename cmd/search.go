package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"mangatool/config"
	"mangatool/core"
	"mangatool/core/sources/mangadex"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for manga",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		lang, _ := cmd.Flags().GetString("lang")
		if lang == "" {
			lang = cfg.DefaultLanguage
		}

		source := mangadex.New()
		results, err := source.Search(args[0], core.SearchOptions{Language: lang, Limit: 20})
		if err != nil {
			return err
		}
		if len(results) == 0 {
			fmt.Println("No results found.")
			return nil
		}
		for _, m := range results {
			fmt.Fprintf(os.Stdout, "[%s] %s (%s)\n", m.ID, m.Title, m.Status)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("lang", "l", "", "language filter (default: from config)")
}
