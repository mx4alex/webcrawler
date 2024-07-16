package main

import (
	"log"
	"webcrawler/internal/crawler"
	"webcrawler/internal/elastic"
	"webcrawler/internal/handler"
	"webcrawler/internal/usecase"
)

func main() {
	esClient, err := elastic.NewElasticsearchClient([]string{"http://localhost:9200"}, "webpages")
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	service := usecase.NewService(esClient)
	h := handler.NewHandler(service)

	crwl := crawler.NewCrawler(service, make(map[string]struct{}), []string{})

	crwl.RunCrawl("https://www.vesti.ru/")

	h.InitRoutes()
}
