/*@ Sourav Dey
Description : It is a script which crawls a web page and puts all the unique URLs into a queue which is then crawled to fetch
all the url on that page. It is multithreaded i.e. goroutines are used to speed up the crawling and one slow page will not 
block the crawling process. Used a mutex lock to print Urls of one thread at a time.
output:  the page being crawled is printed without a tab space and all urls on that page are printed with a tab seperation.
printed start and stop time of each page being crawled to check they are being crawled concurrently.
*/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/jackdanger/collectlinks"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
	"io"
)

var mu = &sync.Mutex{}

func main() {
	flag.Parse()
	args := flag.Args()
	fmt.Println(args)
	if len(args) < 1 {
		fmt.Println("Please specify base URL to crawl ")
		os.Exit(1)
	}
	_, err := url.ParseRequestURI(args[0])
	if err != nil {
		fmt.Println("Please enter a valid URL to crawl")
		os.Exit(2)
	}

	queue := make(chan string)
	filteredQueue := make(chan string)

	go func() { queue <- args[0] }()
	go filterQueue(queue, filteredQueue)

	done := make(chan bool)

	for i := 0; i < 5; i++ {
		go func() {
			for uri := range filteredQueue {
				addToQueue(uri, queue)
			}
			done <- true
		}()
	}
	<-done
}

func filterQueue(in chan string, out chan string) {
	var seen = make(map[string]bool)
	for val := range in {
		if !seen[val] {
			seen[val] = true
			out <- val
		}
	}
}

func addToQueue(uri string, queue chan string) {
	start := time.Now()
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	links := collectlinks.All(resp.Body)
	foundUrls := []string{}
	for _, link := range links {
		absolute := cleanUrl(link, uri)
		foundUrls = append(foundUrls, absolute)
		if uri != "" {
			go func() { queue <- absolute }()
		}
	}
	stop := time.Now()
	display(uri, foundUrls, start, stop)

	// Store the URLs in a text file
	storeURLsInFile(uri, foundUrls)
}

func display(uri string, found []string, start time.Time, stop time.Time) {
	mu.Lock()
	fmt.Println("Start time of crawl of this URL:", start)
	fmt.Println("Stop time of crawl of this URL:", stop)
	fmt.Println(uri)
	for _, str := range found {
		str, err := url.Parse(str)
		if err == nil {
			if str.Scheme == "http" || str.Scheme == "https" {
				fmt.Println("\t", str)
			}
		}
	}
	mu.Unlock()
}

func cleanUrl(href, base string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}
	baseUrl, err := url.Parse(base)
	if err != nil {
		return ""
	}
	uri = baseUrl.ResolveReference(uri)
	return uri.String()
}

func storeURLsInFile(uri string, foundUrls []string) {
	fileName := "urls.txt"
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return
	}
	defer file.Close()

	writer := io.MultiWriter(os.Stdout, file)

	fmt.Fprintln(writer, "URL:", uri)
	for _, u := range foundUrls {
		fmt.Fprintln(writer, "\t", u)
	}
}
