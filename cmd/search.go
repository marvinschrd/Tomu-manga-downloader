package cmd

import (
	"fmt"

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
			fmt.Printf("  %s  No results found for %q\n", yellow("🔍"), args[0])
			return nil
		}
		fmt.Println()
		for _, m := range results {
			fmt.Printf("  📖  %s  %s\n",
				bold(fmt.Sprintf("[%s]", m.ID)),
				m.Title+dim(fmt.Sprintf(" (%s)", m.Status)),
			)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("lang", "l", "", "language filter (default: from config)")
}
