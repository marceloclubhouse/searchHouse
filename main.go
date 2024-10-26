package main

import (
	"flag"
	"searchhouse/spider"
)

func main() {
	// Frontier (pages.db) must be reset if numRoutines changes in between runs!
	numRoutines := flag.Int("numRoutines", 1, "Number of routines for spider to use")
	pageDir := flag.String("pageDir", "pages", "Location for pages to be saved")
	seed := flag.String("seed", "", "First page to start out crawling with")
	maxLinks := flag.Int("maxLinks", 20, "Maximum number of links acceptable within a web page (memory usage)")

	flag.Parse()

	s := spider.NewSpider(*numRoutines, *pageDir, []string{*seed}, *maxLinks)
	s.CrawlConcurrently()
	return
}
