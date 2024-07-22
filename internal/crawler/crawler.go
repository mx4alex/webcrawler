package crawler

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"strings"
	"sync"
	"webcrawler/internal/elastic"
	"webcrawler/internal/entity"
	"webcrawler/internal/extracter"
	"webcrawler/internal/storage"
)

type Crawler struct {
	Logger     *zap.SugaredLogger
	elc        *elastic.ElasticsearchClient
	links      *storage.RedisStorage
	queue      *storage.RedisStorage
	postgres   *storage.PostgresStorage
	mu         sync.Mutex
	countLinks int
}

func NewCrawler(logger *zap.SugaredLogger, elc *elastic.ElasticsearchClient, links *storage.RedisStorage, queue *storage.RedisStorage, page *storage.PostgresStorage) *Crawler {
	return &Crawler{
		Logger:   logger,
		elc:      elc,
		links:    links,
		queue:    queue,
		postgres: page,
	}
}

func (c *Crawler) AddData(startURL string, countKeywords int) error {
	elasticData, pageData, err := c.Crawl(startURL, countKeywords)
	if err != nil {
		c.Logger.Infow("Error in Crawl", err)
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	err = c.elc.IndexDocument(elasticData)
	if err != nil {
		c.Logger.Infow("Error in AddData", err)
		return err
	}

	err = c.postgres.AddPage(pageData)
	if err != nil {
		c.Logger.Infow("Error in AddPage", err)
		return err
	}

	c.countLinks++

	return nil
}

func (c *Crawler) Crawl(url string, countKeywords int) (entity.ElasticData, entity.PageData, error) {
	fmt.Printf("GET URL: %s\n", url)
	c.mu.Lock()
	err := c.links.AddLink(url)
	if err != nil {
		c.Logger.Infow("Error in AddLink", err)
		return entity.ElasticData{}, entity.PageData{}, err
	}
	c.mu.Unlock()

	double := make(map[string]struct{})
	collector := colly.NewCollector()

	pageData := entity.PageData{
		URL: url,
	}
	elasticData := entity.ElasticData{
		URL: url,
	}

	collector.OnHTML("title", func(e *colly.HTMLElement) {
		pageData.Title = e.DOM.Text()
	})

	collector.OnHTML("article, main, .content, .post", func(e *colly.HTMLElement) {
		e.DOM.Find("script, style, nav, footer").Remove()
		pageData.Content = e.DOM.Text()
	})

	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			if _, ok := double[link]; !ok {
				double[link] = struct{}{}

				c.mu.Lock()
				ok, err := c.links.LinkExists(link)
				if err != nil {
					c.Logger.Infow("Error in LinkExists", err)
					return
				}
				if !ok {
					err = c.queue.Push(link)
					if err != nil {
						c.Logger.Infow("Error in Push", err)
						return
					}
				}
				c.mu.Unlock()

			}
		}
	})

	err = collector.Visit(url)
	if err != nil {
		c.Logger.Infow("Error in Visit URL", "url", url, "err", err)
		return entity.ElasticData{}, entity.PageData{}, err
	}

	var text strings.Builder
	text.WriteString(pageData.Title)
	text.WriteString(" ")

	words := extracter.RunRake(pageData.Content)
	for i := 0; i < countKeywords && i < len(words); i++ {
		text.WriteString(words[i].Key)
		text.WriteString(" ")
	}

	newID := uuid.New().String()
	elasticData.ID = newID
	elasticData.Content = text.String()

	pageData.ID = newID

	return elasticData, pageData, nil
}

func (c *Crawler) RunCrawl(startURL string, countWorkers, maxCountLinks, countKeywords int) error {
	exist, err := c.links.LinkExists(startURL)
	if err != nil {
		c.Logger.Infow("Error in LinkExists", err)
		return err
	}
	if !exist {
		err = c.AddData(startURL, countKeywords)
		if err != nil {
			return err
		}
	}

	wg := &sync.WaitGroup{}
	workerInput := make(chan string)

	for i := 0; i < countWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for url := range workerInput {
				c.mu.Lock()
				ok, err := c.links.LinkExists(url)
				if err != nil {
					c.Logger.Infow("Error in LinkExists", err)
					return
				}
				c.mu.Unlock()

				if !ok {
					err = c.AddData(url, countKeywords)
					if err != nil {
						return
					}
				}
			}
		}()
	}

	var nextURL string
	lenght, err := c.queue.Length()
	if err != nil {
		c.Logger.Infow("Error in QueueLength", err)
		return err
	}

	for lenght != 0 && c.countLinks < maxCountLinks {
		nextURL, err = c.queue.Pop()
		if err != nil {
			c.Logger.Infow("Error in QueuePop", err)
			return err
		}

		workerInput <- nextURL

		lenght, err = c.queue.Length()
		if err != nil {
			c.Logger.Infow("Error in QueueLength", err)
			return err
		}
	}

	close(workerInput)
	wg.Wait()

	return nil
}
