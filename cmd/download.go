package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"mangatool/config"
	"mangatool/core"
	"mangatool/core/downloader"
	"mangatool/core/library"
	"mangatool/core/metadata"
	"mangatool/core/sources/mangadex"
)

var (
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	bold   = color.New(color.Bold).SprintFunc()
	dim    = color.New(color.Faint).SprintFunc()
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

		spinner := progressbar.NewOptions(-1,
			progressbar.OptionSetDescription("  Fetching manga info..."),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetWriter(os.Stderr),
		)
		spinner.Add(1)

		manga, err := source.GetManga(mangaID)
		if err != nil {
			spinner.Clear()
			return err
		}
		allChapters, err := source.GetChapters(mangaID, core.ChapterOptions{Language: lang})
		spinner.Clear()
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
				fmt.Printf("  %s  No downloadable chapters found for %q in language %q.\n", yellow("ℹ"), manga.Title, lang)
				fmt.Printf("  %s  Official publisher chapters (Viz, Shonen Jump…) are hosted externally.\n", dim("↳"))
			} else {
				fmt.Printf("  %s  No chapters matched range %q. Available: Ch.%.4g – Ch.%.4g (%d total)\n",
					yellow("ℹ"), chapterRange,
					allChapters[0].Number,
					allChapters[len(allChapters)-1].Number,
					len(allChapters),
				)
			}
			return nil
		}

		fmt.Printf("\n%s  %s\n\n", bold("📚"), bold(fmt.Sprintf("%s  ·  %d chapters available", manga.Title, len(chapters))))

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

		var downloaded, skipped int
		total := len(chapters)

		for i, ch := range chapters {
			cbzName := chapterFilename(manga.Title, ch)
			cbzPath := filepath.Join(mangaDir, cbzName)
			counter := dim(fmt.Sprintf("[%d/%d]", i+1, total))

			if !force {
				if exists, _ := lib.HasChapter(ch.ID); exists {
					fmt.Printf("  %s %s  %s\n", yellow("⏭"), counter, dim(fmt.Sprintf("Ch.%v — already downloaded", ch.Number)))
					skipped++
					continue
				}
			}

			fmt.Printf("  %s %s  Ch.%v\n", "📥", counter, ch.Number)
			tmpDir, err := os.MkdirTemp("", "mangatool-*")
			if err != nil {
				return err
			}

			imageURLs, err := source.DownloadChapter(ch, tmpDir)
			if err != nil {
				os.RemoveAll(tmpDir)
				fmt.Printf("  %s %s  Ch.%v — %s\n", red("⚠"), counter, ch.Number, dim(err.Error()))
				skipped++
				continue
			}
			if len(imageURLs) == 0 {
				os.RemoveAll(tmpDir)
				fmt.Printf("  %s %s  Ch.%v — %s\n", yellow("⏭"), counter, ch.Number, dim("hosted on official publisher site"))
				skipped++
				continue
			}

			bar := progressbar.NewOptions(len(imageURLs),
				progressbar.OptionSetWidth(30),
				progressbar.OptionSetDescription(fmt.Sprintf("      %s", dim("downloading pages"))),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowCount(),
			)

			imagePaths, err := downloader.DownloadImages(imageURLs, tmpDir, func() { bar.Add(1) })
			if err != nil {
				bar.Clear()
				os.RemoveAll(tmpDir)
				return fmt.Errorf("Ch.%v images: %w", ch.Number, err)
			}
			bar.Clear()

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

			size := fileSize(cbzPath)
			fmt.Printf("  %s %s  Ch.%v saved  %s\n", green("✅"), counter, ch.Number, dim("· "+size))
			downloaded++
		}

		fmt.Printf("\n%s\n", dim(strings.Repeat("─", 45)))
		fmt.Printf("  📦  %s downloaded  %s skipped  %s\n",
			bold(fmt.Sprintf("%d", downloaded)),
			dim(fmt.Sprintf("%d", skipped)),
			dim("· "+mangaDir),
		)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.Flags().StringP("chapters", "c", "", "chapter range, e.g. 1-10")
	downloadCmd.Flags().StringP("lang", "l", "", "language (default: from config)")
	downloadCmd.Flags().Bool("force", false, "re-download already downloaded chapters")
}

func fileSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}
	mb := float64(info.Size()) / 1024 / 1024
	if mb < 1 {
		return fmt.Sprintf("%.0f KB", mb*1024)
	}
	return fmt.Sprintf("%.1f MB", mb)
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
