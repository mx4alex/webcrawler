package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"webcrawler/internal/config"
	"webcrawler/internal/crawler"
	"webcrawler/internal/elastic"
	"webcrawler/internal/handler"
	"webcrawler/internal/router"
	"webcrawler/internal/storage"
)

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Error in NewProduction(): %v\n", err)
		return
	}

	defer func(zapLogger *zap.Logger) {
		err = zapLogger.Sync()
		if err != nil {
			fmt.Printf("Error in Sync: %v\n", err)
		}
	}(zapLogger)

	logger := zapLogger.Sugar()

	appConfig, err := config.New()
	if err != nil {
		logger.Infow("Error in config.New()", err)
		return
	}

	elasticConfig := appConfig.Elastic
	esClient, err := elastic.NewElasticsearchClient([]string{elasticConfig.Addr}, elasticConfig.Index)
	if err != nil {
		logger.Infow("Error creating Elasticsearch client", err)
		return
	}

	redisConfig := appConfig.Redis
	urlQueue, err := storage.NewRedisStorage(redisConfig.Addr, redisConfig.Password, redisConfig.QueueKey)
	if err != nil {
		logger.Infow("Error creating URL queue", err)
		return
	}

	urlSet, err := storage.NewRedisStorage(redisConfig.Addr, redisConfig.Password, redisConfig.SetKey)
	if err != nil {
		logger.Infow("Error creating URL set", err)
		return
	}

	pageStorage, err := storage.NewPostgresStorage(appConfig.Postgres)
	if err != nil {
		logger.Infow("Error creating page storage", err)
		return
	}

	h := handler.NewHandler(logger, esClient, pageStorage)

	crwl := crawler.NewCrawler(logger, esClient, urlSet, urlQueue, pageStorage)

	err = crwl.RunCrawl(appConfig.StartURL, appConfig.CountWorkers)
	if err != nil {
		logger.Infow("Error running crawler", err)
		return
	}

	r := router.NewRouter()

	r.Handle("GET", "/webcrawler/search", h.GetUrls)

	logger.Infow("starting server",
		"type", "START",
		"addr", appConfig.HostAddr,
	)

	err = http.ListenAndServe(appConfig.HostAddr, r)
	if err != nil {
		logger.Infow("Error in ListenAndServe", err)
		return
	}
}
