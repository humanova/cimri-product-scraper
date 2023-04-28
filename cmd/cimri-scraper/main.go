package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Config struct {
	Pages        []Page `json:"pages"`
	ProxyURLs    []string `json:"proxy_urls"`
}

type Page struct {
	URL       string `json:"url"`
	PageCount int    `json:"page_count"`
}

func main() {
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %s", err)
	}
	defer configFile.Close()

	var config Config
	err = json.NewDecoder(configFile).Decode(&config)
	if err != nil {
		log.Fatalf("Failed to decode config file: %s", err)
	}

	products := make(chan string)
	uniqueProducts := make(map[string]bool)
	var wg sync.WaitGroup


	for _, page := range config.Pages {
		wg.Add(1)
		go scrapeProductNames(page, config.ProxyURLs, products, &wg)
		time.Sleep(120 * time.Second)
	}

	go func() {
		wg.Wait()
		log.Printf("Scraped %d product names from %d categories\n", len(products), len(config.Pages))
		close(products)
	}()

	// remove duplicates
	for productName := range products {
		uniqueProducts[productName] = true
	}

	outputFile, err := os.Create("product_names.csv")
	if err != nil {
		log.Fatalf("Failed to create file: %s", err)
	}

	writer := csv.NewWriter(outputFile)
	for productName := range uniqueProducts {
		err = writer.Write([]string{productName})
		if err != nil {
			log.Fatalf("Failed to write product name to CSV: %s", err)
		}
	}
	writer.Flush()
	outputFile.Close()

	log.Println("Saved product names to product_names.csv")
}

func scrapeProductNames(page Page, proxyURLs []string, products chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	c := colly.NewCollector(colly.AllowURLRevisit())
	var retryCount uint = 0
	var successfulCount uint = 0
	var retryCountMutex sync.Mutex
	var successfulCountMutex sync.Mutex

	extensions.RandomUserAgent(c)

	// limit the parallelism and add a delay between requests
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 5,
		Delay:       620 * time.Millisecond, // Base delay of 750 ms
		RandomDelay: 1000 * time.Millisecond, // Random delay up to 2000 ms
	})

	c.OnError(func (r *colly.Response, e error) {
		log.Printf("Error (%d) : %s : %s\nUser Agent : %s\nRetrying...\n", r.StatusCode, e.Error(), r.Request.URL.String(),
			r.Request.Headers.Get("User-Agent"))

		retryCountMutex.Lock()
		retryCount++
		log.Printf("Current success/retry count : %d/%d", successfulCount, retryCount)
		retryCountMutex.Unlock()
		time.Sleep(time.Duration(rand.Intn(120)+75) * time.Second)
		r.Request.Retry()
	})

	c.OnResponse(func(r *colly.Response) {
		if r.StatusCode == 200 {
			successfulCountMutex.Lock()
			successfulCount += 1
			successfulCountMutex.Unlock()
		}
	})

	c.OnHTML("div[class^='ProductCard_productName__'] p", func(e *colly.HTMLElement) {
		productName := e.Text
		products <- productName
	})

	for i := 1; i <= page.PageCount; i++ {
		url := fmt.Sprintf("%s?page=%d", page.URL, i)
		c.Visit(url)
	}
}
