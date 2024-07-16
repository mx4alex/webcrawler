package main

import (
	"log"
	"webcrawler/internal/crawler"
	"webcrawler/internal/elastic"
	"webcrawler/internal/handler"
	"webcrawler/internal/storage"
	"webcrawler/internal/usecase"
)

func main() {
	esClient, err := elastic.NewElasticsearchClient([]string{"http://localhost:9200"}, "webpages")
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	urlQueue, err := storage.NewURLQueue("localhost:6379", "", "url_queue")
	if err != nil {
		log.Fatalf("Error creating URL queue: %s", err)
	}

	service := usecase.NewService(esClient)
	h := handler.NewHandler(service)

	crwl := crawler.NewCrawler(service, make(map[string]struct{}), urlQueue)

	crwl.RunCrawl("https://www.vesti.ru/")

	h.InitRoutes()
}
