package usecase

import (
	"encoding/json"
	"fmt"
	"log"
	"webcrawler/internal/elastic"
)

type Service struct {
	elc *elastic.ElasticsearchClient
}

func NewService(elc *elastic.ElasticsearchClient) *Service {
	return &Service{elc}
}

func (s *Service) SearchURLs(searchWords []string) ([]string, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": func() []map[string]interface{} {
					queries := make([]map[string]interface{}, len(searchWords))
					for i, word := range searchWords {
						queries[i] = map[string]interface{}{
							"match": map[string]interface{}{
								"content": word,
							},
						}
					}
					return queries
				}(),
			},
		},
	}

	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("Error encoding query: %v", err)
	}

	urls, err := s.elc.SearchDocument(queryJSON)
	if err != nil {
		return nil, err
	}

	return urls, nil
}

func (s *Service) AddData(elasticData elastic.ElasticData) error {
	err := s.elc.IndexDocument(elasticData)
	if err != nil {
		log.Printf("Error indexing document: %s", err)
		return err
	}

	return nil
}
