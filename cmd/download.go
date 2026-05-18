package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"mangatool/config"
	"mangatool/core"
	"mangatool/core/downloader"
	"mangatool/core/library"
	"mangatool/core/metadata"
	"mangatool/core/sources/mangadex"
)

var downloadCmd = &cobra.Command{
	Use:   "download <manga-id>",
	Short: "Download chapters as CBZ files",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		mangaID := args[0]
		lang, _ := cmd.Flags().GetString("lang")
		if lang == "" {
			lang = cfg.DefaultLanguage
		}
		chapterRange, _ := cmd.Flags().GetString("chapters")
		force, _ := cmd.Flags().GetBool("force")

		source := mangadex.New()
		manga, err := source.GetManga(mangaID)
		if err != nil {
			return err
		}

		allChapters, err := source.GetChapters(mangaID, core.ChapterOptions{Language: lang})
		if err != nil {
			return err
		}

		chapters := allChapters
		if chapterRange != "" {
			chapters, err = filterChapterRange(allChapters, chapterRange)
			if err != nil {
				return err
			}
		}

		if len(chapters) == 0 {
			if len(allChapters) == 0 {
				fmt.Printf("No downloadable chapters found for %q in language %q.\n", manga.Title, lang)
				fmt.Println("Note: official publisher chapters (e.g. Viz, Shonen Jump) are hosted externally and cannot be downloaded.")
			} else {
				fmt.Printf("No chapters matched range %q. Available: Ch.%.4g – Ch.%.4g (%d total)\n",
					chapterRange,
					allChapters[0].Number,
					allChapters[len(allChapters)-1].Number,
					len(allChapters),
				)
			}
			return nil
		}

		lib, err := library.Open(libraryPath())
		if err != nil {
			return err
		}
		defer lib.Close()

		mangaDir := filepath.Join(cfg.OutputDir, sanitize(manga.Title))
		if err := os.MkdirAll(mangaDir, 0755); err != nil {
			return err
		}
		if err := lib.AddManga(manga, mangaDir); err != nil {
			return err
		}

		for _, ch := range chapters {
			cbzName := chapterFilename(manga.Title, ch)
			cbzPath := filepath.Join(mangaDir, cbzName)

			if !force {
				if exists, _ := lib.HasChapter(ch.ID); exists {
					fmt.Printf("  skip Ch.%v (already downloaded — use --force to re-download)\n", ch.Number)
					continue
				}
			}

			fmt.Printf("  downloading Ch.%v...\n", ch.Number)
			tmpDir, err := os.MkdirTemp("", "mangatool-*")
			if err != nil {
				return err
			}

			imageURLs, err := source.DownloadChapter(ch, tmpDir)
			if err != nil {
				os.RemoveAll(tmpDir)
				fmt.Printf("  skip Ch.%v (not available for download: %v)\n", ch.Number, err)
				continue
			}
			if len(imageURLs) == 0 {
				os.RemoveAll(tmpDir)
				fmt.Printf("  skip Ch.%v (no images — hosted on external publisher site)\n", ch.Number)
				continue
			}

			imagePaths, err := downloader.DownloadImages(imageURLs, tmpDir)
			if err != nil {
				os.RemoveAll(tmpDir)
				return fmt.Errorf("Ch.%v images: %w", ch.Number, err)
			}

			ci, err := metadata.GenerateComicInfo(manga, ch, len(imagePaths))
			if err != nil {
				os.RemoveAll(tmpDir)
				return err
			}

			if err := downloader.CreateCBZ(cbzPath, imagePaths, ci); err != nil {
				os.RemoveAll(tmpDir)
				return err
			}
			os.RemoveAll(tmpDir)

			_ = lib.AddChapter(core.DownloadedChapter{
				Chapter:      ch,
				CBZPath:      cbzPath,
				DownloadedAt: time.Now(),
			})
			fmt.Printf("  saved %s\n", cbzName)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("chapters", "c", "", "chapter range, e.g. 1-10")
	downloadCmd.Flags().StringP("lang", "l", "", "language (default: from config)")
	downloadCmd.Flags().Bool("force", false, "re-download already downloaded chapters")
}

func filterChapterRange(chapters []core.Chapter, rangeStr string) ([]core.Chapter, error) {
	parts := strings.SplitN(rangeStr, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid chapter range %q — use format like 1-10", rangeStr)
	}
	from, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid start chapter: %w", err)
	}
	to, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid end chapter: %w", err)
	}
	var out []core.Chapter
	for _, ch := range chapters {
		if ch.Number >= from && ch.Number <= to {
			out = append(out, ch)
		}
	}
	return out, nil
}

func chapterFilename(title string, ch core.Chapter) string {
	if ch.Volume != "" {
		return fmt.Sprintf("%s - Vol.%s - Ch.%s.cbz", sanitize(title), ch.Volume, formatNum(ch.Number))
	}
	return fmt.Sprintf("%s - Ch.%s.cbz", sanitize(title), formatNum(ch.Number))
}

func formatNum(n float64) string {
	if n == float64(int(n)) {
		return fmt.Sprintf("%d", int(n))
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func sanitize(s string) string {
	replacer := strings.NewReplacer("/", "-", "\\", "-", ":", "-", "*", "", "?", "", `"`, "", "<", "", ">", "", "|", "")
	return replacer.Replace(s)
}

func libraryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "mangatool", "library.db")
}
