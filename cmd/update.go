package cmd

import (
	"fmt"
	"os"
	"sync"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"mangatool/config"
	"mangatool/core"
	"mangatool/core/engine"
	"mangatool/core/library"
	"mangatool/core/sources/mangadex"
)

var updateCmd = &cobra.Command{
	Use:   "update [manga-id]",
	Short: "Download new chapters for tracked manga",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if len(args) == 0 && !all {
			return fmt.Errorf("specify a manga ID or use --all to update all tracked manga")
		}

		cfg, _ := config.Load()
		lib, err := library.Open(libraryPath())
		if err != nil {
			return err
		}
		defer lib.Close()

		source := mangadex.New()

		if len(args) == 1 {
			return updateSingle(cmd, args[0], source, lib, cfg.DefaultLanguage)
		}
		return updateAll(cmd, source, lib, cfg.DefaultLanguage)
	},
}

// updateSingle updates one manga — same rich UI as download (spinner + per-chapter progress bar).
func updateSingle(cmd *cobra.Command, mangaID string, source core.Source, lib *library.Library, lang string) error {
	m, err := lib.GetManga(mangaID)
	if err != nil {
		return err
	}

	spinner := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("  Checking for new chapters..."),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetWriter(os.Stderr),
	)
	spinner.Add(1)

	max, _ := lib.MaxChapterNumber(m.ID)
	newChapters, err := source.GetLatestChapters(m.ID, max)
	spinner.Clear()
	if err != nil {
		return err
	}

	var chapters []core.Chapter
	for _, ch := range newChapters {
		if ch.Language == lang || ch.Language == "" {
			chapters = append(chapters, ch)
		}
	}

	if len(chapters) == 0 {
		fmt.Printf("  %s  %s\n\n", green("✅"), bold(m.Title)+dim(" — up to date"))
		return nil
	}

	fmt.Printf("\n%s  %s\n\n", bold("📥"), bold(fmt.Sprintf("%s  ·  %d new chapter(s)", m.Title, len(chapters))))

	var downloaded, skipped int
	var mu sync.Mutex
	var bar *progressbar.ProgressBar

	job := engine.Job{
		Manga: m.Manga,
		Options: engine.JobOptions{
			OutputDir: m.OutputDir,
			Language:  lang,
			Chapters:  chapters,
		},
	}

	eng := &engine.Engine{Source: source, Lib: lib, Conc: 1}
	eng.Run(cmd.Context(), []engine.Job{job}, func(evt engine.Event) {
		mu.Lock()
		defer mu.Unlock()
		counter := dim(fmt.Sprintf("[%d/%d]", evt.ChapterIndex, evt.TotalChapters))
		switch evt.Kind {
		case engine.EvtChapterStarted:
			fmt.Printf("  %s %s  Ch.%v\n", "📥", counter, evt.Chapter.Number)
			bar = progressbar.NewOptions(evt.TotalImages,
				progressbar.OptionSetWidth(30),
				progressbar.OptionSetDescription(fmt.Sprintf("      %s", dim("downloading pages"))),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowCount(),
			)
		case engine.EvtImageProgress:
			if bar != nil {
				bar.Add(1)
			}
		case engine.EvtChapterSaved:
			if bar != nil {
				bar.Clear()
				bar = nil
			}
			size := fileSize(evt.CBZPath)
			fmt.Printf("  %s %s  Ch.%v saved  %s\n", green("✅"), counter, evt.Chapter.Number, dim("· "+size))
			downloaded++
		case engine.EvtChapterSkipped:
			fmt.Printf("  %s %s  %s\n", yellow("⏭"), counter, dim(fmt.Sprintf("Ch.%v — already downloaded", evt.Chapter.Number)))
			skipped++
		case engine.EvtChapterExternal:
			fmt.Printf("  %s %s  %s\n", yellow("⏭"), counter, dim(fmt.Sprintf("Ch.%v — hosted on official publisher site", evt.Chapter.Number)))
			skipped++
		case engine.EvtChapterFailed:
			if bar != nil {
				bar.Clear()
				bar = nil
			}
			fmt.Printf("  %s %s  Ch.%v — %s\n", red("⚠"), counter, evt.Chapter.Number, dim(evt.Err.Error()))
			skipped++
		}
	})

	fmt.Printf("\n  📦  %s downloaded  %s skipped\n\n",
		bold(fmt.Sprintf("%d", downloaded)),
		dim(fmt.Sprintf("%d", skipped)),
	)
	return nil
}

// updateAll updates all tracked manga in parallel (3 concurrent) with log-style output.
func updateAll(cmd *cobra.Command, source core.Source, lib *library.Library, lang string) error {
	spinner := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription("  Loading library..."),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetWriter(os.Stderr),
	)
	spinner.Add(1)

	targets, err := lib.ListManga()
	spinner.Clear()
	if err != nil {
		return err
	}

	if len(targets) == 0 {
		fmt.Printf("  %s  Library is empty. Use 'download' to add manga.\n\n", dim("ℹ"))
		return nil
	}

	fmt.Printf("\n  Updating %s manga...\n\n", bold(fmt.Sprintf("%d", len(targets))))

	jobs := make([]engine.Job, len(targets))
	for i, m := range targets {
		jobs[i] = engine.Job{
			Manga: m.Manga,
			Options: engine.JobOptions{
				OutputDir: m.OutputDir,
				Language:  lang,
				// Chapters: nil → engine fetches new chapters internally
			},
		}
	}

	var mu sync.Mutex
	eng := &engine.Engine{Source: source, Lib: lib, Conc: 3}
	eng.Run(cmd.Context(), jobs, func(evt engine.Event) {
		mu.Lock()
		defer mu.Unlock()
		switch evt.Kind {
		case engine.EvtChapterSaved:
			fmt.Printf("  %s  %-30s  Ch.%v saved  %s\n",
				"📥", bold(evt.MangaTitle), evt.Chapter.Number, dim("· "+fileSize(evt.CBZPath)))
		case engine.EvtChapterFailed:
			fmt.Printf("  %s  %-30s  Ch.%v — %s\n",
				red("⚠"), evt.MangaTitle, evt.Chapter.Number, dim(evt.Err.Error()))
		case engine.EvtMangaDone:
			if evt.Err != nil {
				fmt.Printf("  %s  %-30s  %s\n", red("⚠"), evt.MangaTitle, dim(evt.Err.Error()))
			} else if evt.Downloaded > 0 {
				fmt.Printf("  %s  %-30s  %s\n",
					"📥", bold(evt.MangaTitle), green(fmt.Sprintf("%d new chapter(s)", evt.Downloaded)))
			} else {
				fmt.Printf("  %s  %-30s  %s\n",
					green("✅"), bold(evt.MangaTitle), dim("up to date"))
			}
		}
	})

	fmt.Println()
	return nil
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().Bool("all", false, "update all tracked manga (3 concurrent)")
}
