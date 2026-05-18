package mangadex

type mangaListResponse struct {
	Result string      `json:"result"`
	Data   []mangaData `json:"data"`
	Total  int         `json:"total"`
}

type mangaResponse struct {
	Result string    `json:"result"`
	Data   mangaData `json:"data"`
}

type mangaData struct {
	ID            string          `json:"id"`
	Attributes    mangaAttributes `json:"attributes"`
	Relationships []relationship  `json:"relationships"`
}

type mangaAttributes struct {
	Title       map[string]string `json:"title"`
	Description map[string]string `json:"description"`
	Status      string            `json:"status"`
}

type chapterFeedResponse struct {
	Result string        `json:"result"`
	Data   []chapterData `json:"data"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

type chapterData struct {
	ID         string            `json:"id"`
	Attributes chapterAttributes `json:"attributes"`
}

type chapterAttributes struct {
	Title              string `json:"title"`
	Volume             string `json:"volume"`
	Chapter            string `json:"chapter"`
	TranslatedLanguage string `json:"translatedLanguage"`
	Pages              int    `json:"pages"`
	ExternalUrl        string `json:"externalUrl"` // non-empty = official publisher chapter, no images on MangaDex
}

type atHomeResponse struct {
	Result  string `json:"result"`
	BaseURL string `json:"baseUrl"`
	Chapter struct {
		Hash string   `json:"hash"`
		Data []string `json:"data"`
	} `json:"chapter"`
}

type relationship struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
