package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"mangatool/config"
	"mangatool/core"
	"mangatool/core/downloader"
	"mangatool/core/library"
	"mangatool/core/metadata"
	"mangatool/core/sources/mangadex"
)

var updateCmd = &cobra.Command{
	Use:   "update [manga-id]",
	Short: "Download new chapters for tracked manga",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _ := config.Load()
		lib, err := library.Open(libraryPath())
		if err != nil {
			return err
		}
		defer lib.Close()

		var targets []core.DownloadedManga
		if len(args) == 1 {
			m, err := lib.GetManga(args[0])
			if err != nil {
				return err
			}
			targets = []core.DownloadedManga{m}
		} else {
			targets, err = lib.ListManga()
			if err != nil {
				return err
			}
		}

		source := mangadex.New()
		for _, m := range targets {
			maxNum, err := lib.MaxChapterNumber(m.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "skipping %s: %v\n", m.Title, err)
				continue
			}

			newChapters, err := source.GetLatestChapters(m.ID, maxNum)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error checking %s: %v\n", m.Title, err)
				continue
			}

			var filtered []core.Chapter
			for _, ch := range newChapters {
				if ch.Language == cfg.DefaultLanguage || ch.Language == "" {
					filtered = append(filtered, ch)
				}
			}

			if len(filtered) == 0 {
				fmt.Printf("%s: up to date\n", m.Title)
				continue
			}
			fmt.Printf("%s: %d new chapter(s)\n", m.Title, len(filtered))

			for _, ch := range filtered {
				fmt.Printf("  downloading Ch.%v...\n", ch.Number)
				tmpDir, _ := os.MkdirTemp("", "mangatool-*")

				imageURLs, err := source.DownloadChapter(ch, tmpDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Ch.%v failed: %v\n", ch.Number, err)
					os.RemoveAll(tmpDir)
					continue
				}

				imagePaths, err := downloader.DownloadImages(imageURLs, tmpDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  Ch.%v images failed: %v\n", ch.Number, err)
					os.RemoveAll(tmpDir)
					continue
				}

				ci, _ := metadata.GenerateComicInfo(m.Manga, ch, len(imagePaths))
				cbzPath := filepath.Join(m.OutputDir, chapterFilename(m.Title, ch))
				if err := downloader.CreateCBZ(cbzPath, imagePaths, ci); err != nil {
					fmt.Fprintf(os.Stderr, "  Ch.%v CBZ failed: %v\n", ch.Number, err)
					os.RemoveAll(tmpDir)
					continue
				}
				os.RemoveAll(tmpDir)

				_ = lib.AddChapter(core.DownloadedChapter{
					Chapter:      ch,
					CBZPath:      cbzPath,
					DownloadedAt: time.Now(),
				})
				fmt.Printf("  saved Ch.%v\n", ch.Number)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
