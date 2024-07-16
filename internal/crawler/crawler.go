package crawler

import (
	"fmt"
	"github.com/gocolly/colly/v2"
	"strings"
	"sync"
	"webcrawler/internal/elastic"
	"webcrawler/internal/extracter"
	"webcrawler/internal/usecase"
)

type Crawler struct {
	service *usecase.Service
	links   map[string]struct{}
	queue   []string
	mu      sync.Mutex
}

func NewCrawler(service *usecase.Service, links map[string]struct{}, queue []string) *Crawler {
	return &Crawler{
		service: service,
		links:   links,
		queue:   queue,
	}
}

type PageData struct {
	URL     string
	Title   string
	Content string
}

func (c *Crawler) RunCrawl(startURL string) error {
	data := c.Crawl(startURL)
	err := c.service.AddData(data)
	if err != nil {
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
				_, ok := c.links[url]
				c.mu.Unlock()
				if !ok {
					data := c.Crawl(url)
					err := c.service.AddData(data)
					if err != nil {
						fmt.Println(err)
						return
					}
				}
			}
		}()
	}

	var nextURL string
	for len(c.queue) != 0 && len(c.queue) < 10000 {
		nextURL = c.queue[0]
		workerInput <- nextURL

		c.queue = c.queue[1:]
	}

	close(workerInput)
	wg.Wait()

	return nil
}

func (c *Crawler) Crawl(url string) elastic.ElasticData {
	fmt.Printf("GET URL: %s\n", url)
	c.mu.Lock()
	c.links[url] = struct{}{}
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
				if _, ok := c.links[link]; !ok {
					c.queue = append(c.queue, link)
				}
				c.mu.Unlock()

			}
		}
	})

	err := collector.Visit(url)
	if err != nil {
		fmt.Errorf("Error in Visit URL %s: %s", url, err)
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

	return elasticData
}
