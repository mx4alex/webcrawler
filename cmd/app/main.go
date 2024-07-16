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

	urlQueue, err := storage.NewRedisStorage("localhost:6379", "", "url_queue")
	if err != nil {
		log.Fatalf("Error creating URL queue: %s", err)
	}

	urlSet, err := storage.NewRedisStorage("localhost:6379", "", "url_set")
	if err != nil {
		log.Fatalf("Error creating URL set: %s", err)
	}

	service := usecase.NewService(esClient)
	h := handler.NewHandler(service)

	crwl := crawler.NewCrawler(service, urlSet, urlQueue)

	crwl.RunCrawl("https://www.vesti.ru/")

	h.InitRoutes()
}
