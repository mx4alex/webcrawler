package crawler

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
	"sync"
	"webcrawler/internal/elastic"
	"webcrawler/internal/extracter"
	"webcrawler/internal/storage"
)

type Crawler struct {
	elc   *elastic.ElasticsearchClient
	links *storage.RedisStorage
	queue *storage.RedisStorage
	mu    sync.Mutex
}

func NewCrawler(elc *elastic.ElasticsearchClient, links *storage.RedisStorage, queue *storage.RedisStorage) *Crawler {
	return &Crawler{
		elc:   elc,
		links: links,
		queue: queue,
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
		fmt.Printf("Error in Crawl: %v", err)
		return err
	}

	err = c.elc.IndexDocument(data)
	if err != nil {
		fmt.Printf("Error in AddData: %v", err)
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
					fmt.Printf("Error in LinkExists: %v", err)
					return
				}
				c.mu.Unlock()

				if !ok {
					data, err := c.Crawl(url)
					if err != nil {
						fmt.Printf("Error in Crawl: %v", err)
						return
					}
					err = c.elc.IndexDocument(data)
					if err != nil {
						fmt.Printf("Error in AddData: %v", err)
						return
					}
				}
			}
		}()
	}

	var nextURL string
	lenght, err := c.queue.Length()
	if err != nil {
		return err
	}

	for lenght != 0 && lenght < 1000 {
		nextURL, err = c.queue.Pop()
		if err != nil {
			return err
		}

		workerInput <- nextURL

		lenght, err = c.queue.Length()
		if err != nil {
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
		fmt.Printf("Error in AddLink: %v", err)
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
					fmt.Printf("Error in LinkExists: %v", err)
					return
				}
				if !ok {
					err = c.queue.Push(link)
					if err != nil {
						fmt.Printf("Error in Push: %v", err)
						return
					}
				}
				c.mu.Unlock()

			}
		}
	})

	err = collector.Visit(url)
	if err != nil {
		fmt.Printf("Error in Visit URL %s: %s", url, err)
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
