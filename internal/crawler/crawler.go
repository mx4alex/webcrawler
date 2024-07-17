package crawler

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"go.uber.org/zap"
	"strings"
	"sync"
	"webcrawler/internal/elastic"
	"webcrawler/internal/extracter"
	"webcrawler/internal/storage"
)

type Crawler struct {
	Logger *zap.SugaredLogger
	elc    *elastic.ElasticsearchClient
	links  *storage.RedisStorage
	queue  *storage.RedisStorage
	mu     sync.Mutex
}

func NewCrawler(logger *zap.SugaredLogger, elc *elastic.ElasticsearchClient, links *storage.RedisStorage, queue *storage.RedisStorage) *Crawler {
	return &Crawler{
		Logger: logger,
		elc:    elc,
		links:  links,
		queue:  queue,
	}
}

type PageData struct {
	URL     string
	Title   string
	Content string
}

func (c *Crawler) RunCrawl(startURL string) error {
	data, err := c.Crawl(startURL)
	if err != nil {
		c.Logger.Infow("Error in Crawl", err)
		return err
	}

	err = c.elc.IndexDocument(data)
	if err != nil {
		c.Logger.Infow("Error in AddData", err)
		return err
	}

	wg := &sync.WaitGroup{}
	workerInput := make(chan string)

	for i := 0; i < 50; i++ {
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
					data, err := c.Crawl(url)
					if err != nil {
						c.Logger.Infow("Error in Crawl", err)
						return
					}
					err = c.elc.IndexDocument(data)
					if err != nil {
						c.Logger.Infow("Error in AddData", err)
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

	for lenght != 0 && lenght < 1000 {
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

func (c *Crawler) Crawl(url string) (elastic.ElasticData, error) {
	fmt.Printf("GET URL: %s\n", url)
	c.mu.Lock()
	err := c.links.AddLink(url)
	if err != nil {
		c.Logger.Infow("Error in AddLink", err)
		return elastic.ElasticData{}, err
	}
	c.mu.Unlock()

	double := make(map[string]struct{})
	collector := colly.NewCollector()

	pageData := PageData{
		URL: url,
	}
	elasticData := elastic.ElasticData{
		URL: url,
	}

	collector.OnHTML("title", func(e *colly.HTMLElement) {
		pageData.Title = e.Text
	})

	collector.OnHTML("body", func(e *colly.HTMLElement) {
		pageData.Content = e.Text
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
		return elastic.ElasticData{}, err
	}

	var text strings.Builder
	text.WriteString(pageData.Title)
	text.WriteString(" ")

	words := extracter.RunRake(pageData.Content)
	for i := 0; i < 50 && i < len(words); i++ {
		text.WriteString(words[i].Key)
		text.WriteString(" ")
	}

	elasticData.Content = text.String()

	return elasticData, nil
}
