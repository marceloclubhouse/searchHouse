package main

import (
	"flag"
	"log"
	"os"
	"searchHouse/spider"
)

func main() {
	// Create a log file
	logFile, err := os.OpenFile("searchHouse.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Set log output to the file
	log.SetOutput(logFile)

	// Arguments for spider
	var isSpider bool
	flag.BoolVar(&isSpider, "spider", false, "Run the spider")
	numRoutines := flag.Int("numRoutines", 1, "Number of routines for spider to use")
	pageDir := flag.String("pageDir", "pages", "Location for pages to be saved")
	seed := flag.String("seed", "", "First page to start out crawling with")
	maxLinks := flag.Int("maxLinks", 20, "Maximum number of links acceptable within a web page (memory usage)")

	flag.Parse()

	if isSpider {
		// Frontier (pages.db) must be reset if numRoutines changes in between runs!
		s := spider.NewSpider(*numRoutines, *pageDir, []string{*seed}, *maxLinks)
		s.CrawlConcurrently()
	}
}
