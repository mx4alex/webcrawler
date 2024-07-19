package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"strings"
	"webcrawler/internal/entity"
)

type ElasticsearchClient struct {
	client *elasticsearch.Client
	index  string
}

func NewElasticsearchClient(urls []string, index string) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: urls,
	}
	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &ElasticsearchClient{client: es, index: index}, nil
}

func (es *ElasticsearchClient) IndexDocument(doc entity.ElasticData) error {
	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("Error marshalling doc: %v", err)
	}

	req := esapi.IndexRequest{
		Index:      es.index,
		DocumentID: doc.ID,
		Body:       bytes.NewReader(body),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("error indexing document: %s", res.String())
	}

	return nil
}

func (es *ElasticsearchClient) SearchDocument(searchWords []string) ([]entity.ElasticData, error) {
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
		return nil, fmt.Errorf("Error marshalling query: %v", err)
	}

	req := esapi.SearchRequest{
		Index: []string{es.index},
		Body:  strings.NewReader(string(queryJSON)),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("Error executing Elasticsearch search: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch error: %v", err)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Error decoding Elasticsearch response: %v", err)
	}

	var searchResults []entity.ElasticData
	hits, found := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if found {
		for _, hit := range hits {
			hitMap := hit.(map[string]interface{})
			source := hitMap["_source"].(map[string]interface{})
			id := hitMap["_id"].(string)
			url := source["url"].(string)
			searchResults = append(searchResults, entity.ElasticData{
				ID:  id,
				URL: url,
			})
		}
	}

	return searchResults, nil
}
