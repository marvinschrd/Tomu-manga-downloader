package mangadex

import (
	"fmt"
	"net/url"
	"strconv"

	"mangatool/core"
	"mangatool/core/httpclient"
)

const defaultBaseURL = "https://api.mangadex.org"

type Source struct {
	client *httpclient.Client
}

func New() *Source {
	return &Source{client: httpclient.New(defaultBaseURL)}
}

// NewWithBaseURL points the source at a custom URL, used in tests.
func NewWithBaseURL(baseURL string) *Source {
	return &Source{client: httpclient.New(baseURL)}
}

func (s *Source) Search(query string, opts core.SearchOptions) ([]core.Manga, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}
	params := url.Values{
		"title":      {query},
		"limit":      {strconv.Itoa(limit)},
		"includes[]": {"cover_art"},
	}
	var resp mangaListResponse
	if err := s.client.Get("/manga", params, &resp); err != nil {
		return nil, err
	}
	return convertMangaList(resp.Data), nil
}

func (s *Source) GetManga(id string) (core.Manga, error) {
	params := url.Values{"includes[]": {"cover_art"}}
	var resp mangaResponse
	if err := s.client.Get("/manga/"+id, params, &resp); err != nil {
		return core.Manga{}, err
	}
	return convertManga(resp.Data), nil
}

func (s *Source) GetChapters(mangaID string, opts core.ChapterOptions) ([]core.Chapter, error) {
	return s.fetchChapters(mangaID, opts.Language, 0)
}

func (s *Source) GetLatestChapters(mangaID string, afterChapterNumber float64) ([]core.Chapter, error) {
	all, err := s.fetchChapters(mangaID, "", 0)
	if err != nil {
		return nil, err
	}
	var latest []core.Chapter
	for _, ch := range all {
		if ch.Number > afterChapterNumber {
			latest = append(latest, ch)
		}
	}
	return latest, nil
}

func (s *Source) DownloadChapter(chapter core.Chapter, destDir string) ([]string, error) {
	var resp atHomeResponse
	if err := s.client.Get("/at-home/server/"+chapter.ID, nil, &resp); err != nil {
		return nil, err
	}
	urls := make([]string, len(resp.Chapter.Data))
	for i, filename := range resp.Chapter.Data {
		urls[i] = fmt.Sprintf("%s/data/%s/%s", resp.BaseURL, resp.Chapter.Hash, filename)
	}
	return urls, nil
}

func (s *Source) fetchChapters(mangaID, language string, offset int) ([]core.Chapter, error) {
	params := url.Values{
		"limit":          {"500"},
		"offset":         {strconv.Itoa(offset)},
		"order[chapter]": {"asc"},
	}
	if language != "" {
		params["translatedLanguage[]"] = []string{language}
	}

	var resp chapterFeedResponse
	if err := s.client.Get("/manga/"+mangaID+"/feed", params, &resp); err != nil {
		return nil, err
	}

	chapters := convertChapterList(mangaID, resp.Data)

	if offset+resp.Limit < resp.Total {
		more, err := s.fetchChapters(mangaID, language, offset+resp.Limit)
		if err != nil {
			return nil, err
		}
		chapters = append(chapters, more...)
	}
	return chapters, nil
}

func convertMangaList(data []mangaData) []core.Manga {
	out := make([]core.Manga, len(data))
	for i, d := range data {
		out[i] = convertManga(d)
	}
	return out
}

func convertManga(d mangaData) core.Manga {
	title := d.Attributes.Title["en"]
	if title == "" {
		for _, v := range d.Attributes.Title {
			title = v
			break
		}
	}
	return core.Manga{
		ID:          d.ID,
		Source:      "mangadex",
		Title:       title,
		Description: d.Attributes.Description["en"],
		Status:      d.Attributes.Status,
	}
}

func convertChapterList(mangaID string, data []chapterData) []core.Chapter {
	out := make([]core.Chapter, 0, len(data))
	for _, d := range data {
		n, err := strconv.ParseFloat(d.Attributes.Chapter, 64)
		if err != nil {
			continue
		}
		out = append(out, core.Chapter{
			ID:        d.ID,
			MangaID:   mangaID,
			Source:    "mangadex",
			Number:    n,
			Title:     d.Attributes.Title,
			Volume:    d.Attributes.Volume,
			Language:  d.Attributes.TranslatedLanguage,
			PageCount: d.Attributes.Pages,
		})
	}
	return out
}
