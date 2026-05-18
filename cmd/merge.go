package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"mangatool/core/library"
	"mangatool/core/merger"
	"mangatool/core/metadata"
	"mangatool/core/sources/mangadex"
)

var mergeCmd = &cobra.Command{
	Use:   "merge <manga-id>",
	Short: "Merge chapters into volumes",
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

		dbChapters, err := lib.GetChapters(m.ID)
		if err != nil {
			return err
		}

		chapterRange, _ := cmd.Flags().GetString("chapters")
		name, _ := cmd.Flags().GetString("name")
		removeOriginals, _ := cmd.Flags().GetBool("remove-originals")

		source := mangadex.New()
		manga, _ := source.GetManga(m.ID)

		entries := make([]merger.ChapterEntry, 0, len(dbChapters))
		for _, ch := range dbChapters {
			entries = append(entries, merger.ChapterEntry{
				Number:  ch.Number,
				Volume:  ch.Volume,
				CBZPath: ch.CBZPath,
			})
		}

		if chapterRange != "" {
			if name == "" {
				name = fmt.Sprintf("Chapters %s", chapterRange)
			}
			filtered, err := filterChapterRangeEntries(entries, chapterRange)
			if err != nil {
				return err
			}
			if len(filtered) == 0 {
				return fmt.Errorf("no downloaded chapters found in range %s", chapterRange)
			}
			paths := entryCBZPaths(filtered)
			outPath := filepath.Join(m.OutputDir, sanitize(fmt.Sprintf("%s - %s.cbz", m.Title, name)))
			ci, _ := metadata.GenerateMergedComicInfo(manga, name, filtered[0].Number, filtered[len(filtered)-1].Number, 0)
			if err := merger.MergeCBZs(paths, outPath, ci); err != nil {
				return err
			}
			fmt.Printf("Created: %s\n", outPath)
			if removeOriginals {
				for _, p := range paths {
					os.Remove(p)
				}
			}
			return nil
		}

		groups := merger.GroupByVolume(entries)
		if len(groups) == 0 {
			return fmt.Errorf("no volume information found — use -c and -n for manual merge")
		}
		for vol, volChapters := range groups {
			paths := entryCBZPaths(volChapters)
			outPath := filepath.Join(m.OutputDir, sanitize(fmt.Sprintf("%s - Vol.%s.cbz", m.Title, vol)))
			ci, _ := metadata.GenerateMergedComicInfo(manga, "Vol."+vol, volChapters[0].Number, volChapters[len(volChapters)-1].Number, 0)
			if err := merger.MergeCBZs(paths, outPath, ci); err != nil {
				fmt.Fprintf(os.Stderr, "Vol.%s failed: %v\n", vol, err)
				continue
			}
			fmt.Printf("Created: %s\n", outPath)
			if removeOriginals {
				for _, p := range paths {
					os.Remove(p)
				}
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringP("chapters", "c", "", "chapter range to merge manually, e.g. 1-10")
	mergeCmd.Flags().StringP("name", "n", "", "name for manually merged file")
	mergeCmd.Flags().Bool("remove-originals", false, "delete original chapter CBZs after merging")
}

func filterChapterRangeEntries(entries []merger.ChapterEntry, rangeStr string) ([]merger.ChapterEntry, error) {
	var from, to float64
	if _, err := fmt.Sscanf(rangeStr, "%f-%f", &from, &to); err != nil {
		return nil, fmt.Errorf("invalid range %q — use format like 1-10", rangeStr)
	}
	var out []merger.ChapterEntry
	for _, e := range entries {
		if e.Number >= from && e.Number <= to {
			out = append(out, e)
		}
	}
	return out, nil
}

func entryCBZPaths(entries []merger.ChapterEntry) []string {
	paths := make([]string, len(entries))
	for i, e := range entries {
		paths[i] = e.CBZPath
	}
	return paths
}
