package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"webcrawler/internal/elastic"
	"webcrawler/internal/router"
)

type Handler struct {
	elc *elastic.ElasticsearchClient
}

func NewHandler(elc *elastic.ElasticsearchClient) *Handler {
	return &Handler{elc}
}

type Response struct {
	URLs []string `json:"urls"`
}

func (h *Handler) InitRoutes() {
	r := router.NewRouter()

	r.Handle("GET", "/webcrawler/search", h.getUrls)

	fmt.Println("Server is running at :8080")
	http.ListenAndServe(":8080", r)
}

func (h *Handler) getUrls(w http.ResponseWriter, r *http.Request) {
	args := strings.Split(r.URL.Path, "/")
	if len(args) != 4 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	words := args[3]

	searchWords := strings.Split(words, " ")

	fmt.Println(searchWords)

	urls, err := h.elc.SearchDocument(searchWords)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error SearchURLs: %v", err), http.StatusInternalServerError)
		return
	}

	response := Response{URLs: urls}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
		return
	}
}
