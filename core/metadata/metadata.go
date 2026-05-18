package metadata

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"mangatool/core"
)

type ComicInfo struct {
	XMLName   xml.Name `xml:"ComicInfo"`
	Series    string   `xml:"Series"`
	Title     string   `xml:"Title,omitempty"`
	Number    string   `xml:"Number,omitempty"`
	Volume    int      `xml:"Volume,omitempty"`
	Summary   string   `xml:"Summary,omitempty"`
	PageCount int      `xml:"PageCount"`
	Language  string   `xml:"LanguageISO,omitempty"`
	Notes     string   `xml:"Notes,omitempty"`
}

func GenerateComicInfo(manga core.Manga, ch core.Chapter, pageCount int) ([]byte, error) {
	ci := ComicInfo{
		Series:    manga.Title,
		Title:     ch.Title,
		Number:    formatChapterNumber(ch.Number),
		PageCount: pageCount,
		Language:  ch.Language,
		Notes:     fmt.Sprintf("Source: %s", manga.Source),
	}
	if ch.Volume != "" {
		if v, err := strconv.Atoi(ch.Volume); err == nil {
			ci.Volume = v
		}
	}
	return marshal(ci)
}

func GenerateMergedComicInfo(manga core.Manga, volumeLabel string, firstChapter, lastChapter float64, pageCount int) ([]byte, error) {
	ci := ComicInfo{
		Series:    manga.Title,
		Title:     volumeLabel,
		Number:    fmt.Sprintf("%v-%v", firstChapter, lastChapter),
		PageCount: pageCount,
		Notes:     fmt.Sprintf("Source: %s", manga.Source),
	}
	return marshal(ci)
}

func formatChapterNumber(n float64) string {
	if n == float64(int(n)) {
		return fmt.Sprintf("%d", int(n))
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

func marshal(ci ComicInfo) ([]byte, error) {
	out, err := xml.MarshalIndent(ci, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), out...), nil
}
