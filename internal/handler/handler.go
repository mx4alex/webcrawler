package handler

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"webcrawler/internal/elastic"
	"webcrawler/internal/storage"
)

type Handler struct {
	Logger   *zap.SugaredLogger
	elc      *elastic.ElasticsearchClient
	postgres *storage.PostgresStorage
}

func NewHandler(logger *zap.SugaredLogger, elc *elastic.ElasticsearchClient, postgres *storage.PostgresStorage) *Handler {
	return &Handler{
		Logger:   logger,
		elc:      elc,
		postgres: postgres,
	}
}

func (h *Handler) GetUrls(w http.ResponseWriter, r *http.Request) {
	args := strings.Split(r.URL.Path, "/")
	if len(args) != 4 {
		h.Logger.Infow("Not correct parameters", "path", r.URL.Path)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	words := args[3]

	searchWords := strings.Split(words, " ")

	fmt.Println(searchWords)

	elasticData, err := h.elc.SearchDocument(searchWords)
	if err != nil {
		h.Logger.Infow("Error in Search Document", err)
		http.Error(w, fmt.Sprintf("Error SearchURLs: %v", err), http.StatusInternalServerError)
		return
	}

	var response []Response
	for _, data := range elasticData {
		pageData, err := h.postgres.GetPage(data.ID)

		if err != nil {
			h.Logger.Infow("Error in GetPage", "ID", data.ID, "error", err)
			http.Error(w, fmt.Sprintf("Error GetPage: %v", err), http.StatusInternalServerError)
			return
		}

		subtext := SearchContext(pageData.Content, words, 5)
		response = append(response, Response{
			URL:  data.URL,
			Text: subtext,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		h.Logger.Infow("Error in Encode Response", err)
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}
