# ConcurrentCrawler
It is a script which crawls a web page and puts all the unique URLs into a queue which is then crawled to fetch all the url on that page. It is multithreaded i.e. goroutines are used to speed up the crawling and one slow page will not  block the crawling process. Used a mutex lock to print Urls of one thread at a time.
