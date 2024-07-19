package entity

type PageData struct {
	ID      string
	URL     string
	Title   string
	Content string
}

type ElasticData struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	Content string `json:"content"`
}
