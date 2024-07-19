package storage

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"webcrawler/internal/config"
	"webcrawler/internal/entity"
)

type PostgresStorage struct {
	DB *sql.DB
}

func NewPostgresStorage(cfg config.PostgresConfig) (*PostgresStorage, error) {
	conn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS pages (id TEXT, url TEXT, title TEXT, content TEXT)")
	if err != nil {
		return nil, err
	}

	return &PostgresStorage{db}, nil
}

func (p *PostgresStorage) AddPage(page entity.PageData) error {
	query := "INSERT INTO pages (id, url, title, content) VALUES ($1, $2, $3, $4)"
	_, err := p.DB.Exec(query, page.ID, page.URL, page.Title, page.Content)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresStorage) GetPage(id string) (entity.PageData, error) {
	var page entity.PageData
	query := "SELECT id, url, title, content FROM pages WHERE id = $1"
	row := p.DB.QueryRow(query, id)
	err := row.Scan(&page.ID, &page.URL, &page.Title, &page.Content)
	return page, err
}
