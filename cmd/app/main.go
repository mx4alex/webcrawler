package main

import (
	"log"
	"webcrawler/internal/config"
	"webcrawler/internal/crawler"
	"webcrawler/internal/elastic"
	"webcrawler/internal/handler"
	"webcrawler/internal/storage"
)

func main() {
	appConfig, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	elasticConfig := appConfig.Elastic
	esClient, err := elastic.NewElasticsearchClient([]string{elasticConfig.Addr}, elasticConfig.Index)
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %s", err)
	}

	redisConfig := appConfig.Redis
	urlQueue, err := storage.NewRedisStorage(redisConfig.Addr, redisConfig.Password, redisConfig.QueueKey)
	if err != nil {
		log.Fatalf("Error creating URL queue: %s", err)
	}

	urlSet, err := storage.NewRedisStorage(redisConfig.Addr, redisConfig.Password, redisConfig.SetKey)
	if err != nil {
		log.Fatalf("Error creating URL set: %s", err)
	}

	h := handler.NewHandler(esClient)

	crwl := crawler.NewCrawler(esClient, urlSet, urlQueue)

	crwl.RunCrawl(appConfig.StartURL)

	h.InitRoutes()
}
