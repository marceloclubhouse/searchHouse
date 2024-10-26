package spider

import (
	"errors"
	lru "github.com/hashicorp/golang-lru/v2"
	"hash/fnv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"searchHouse/common"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type SearchHouseSpider struct {
	numRoutines      int
	frontier         Frontier
	workingDirectory string
	maxLinksPerPage  int
	ioMu             *sync.Mutex
	wordpressSites   *lru.Cache[string, bool]
}

func NewSpider(numRoutines int, workingDirectory string, seed []string, maxLinks int) *SearchHouseSpider {
	ioMu := new(sync.Mutex)
	wpCache, _ := lru.New[string, bool](1000)
	cs := SearchHouseSpider{
		numRoutines:      numRoutines,
		workingDirectory: workingDirectory,
		maxLinksPerPage:  maxLinks,
		ioMu:             ioMu,
		wordpressSites:   wpCache,
	}
	cs.frontier.Init()
	cs.setSeed(seed)
	return &cs
}

func (s *SearchHouseSpider) CrawlConcurrently() {
	wg := new(sync.WaitGroup)
	wg.Add(s.numRoutines)
	for i := 0; i < s.numRoutines; i++ {
		go s.Crawl(i, wg)
	}
	wg.Wait()
}

func (s *SearchHouseSpider) Crawl(routineNum int, wg *sync.WaitGroup) {
	defer wg.Done()
	fp := common.NewFingerprints(3, 10000)
	for {
		currentUrl := s.frontier.PopURL(routineNum)
		if currentUrl == "" {
			time.Sleep(time.Second)
			continue
		}
		if !s.urlValid(currentUrl) {
			continue
		}
		if !s.pageDownloaded(currentUrl) {
			resp, err := http.Get(currentUrl)
			if err == nil && resp.Status == "200 OK" {
				body, err := io.ReadAll(resp.Body)
				if err == nil {
					page := common.NewWebPage(time.Now().Unix(), currentUrl, resp.Status, string(body))
					resp.Body.Close()
					if !s.validPage(page) || s.duplicateExists(fp, page) {
						continue
					}
					fp.InsertFingerprintsUsingWebpage(page)
					s.writeToDisk(*page)
					anchors := s.constructProperURLs(page.FindAllAnchorHREFs(s.maxLinksPerPage), currentUrl)
					for key := range anchors.m {
						if !s.pageDownloaded(key) {
							s.frontier.InsertPage(key, s.calcWebsiteToRoutineNum(key))
						}
					}
				}
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *SearchHouseSpider) writeToDisk(w common.WebPage) {
	fileName := filepath.Join(s.workingDirectory, strconv.FormatUint(s.hash(w.Url), 10)+".json")
	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	_, err = f.Write(w.Serialize())
	if err != nil {
		log.Fatalln(err)
	}
}

func (s *SearchHouseSpider) fileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

func (s *SearchHouseSpider) pageDownloaded(url string) bool {
	fileName := filepath.Join(s.workingDirectory, strconv.FormatUint(s.hash(url), 10)+".json")
	s.ioMu.Lock()
	defer s.ioMu.Unlock()
	exists, err := s.fileExists(fileName)
	if err != nil {
		log.Fatalln(err)
	}
	return exists
}

func (s *SearchHouseSpider) urlValid(url string) bool {
	urlRe := regexp.MustCompile(`^(https://[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}[-a-zA-Z0-9()@:_+~?=/]*)$`)
	extRe := regexp.MustCompile(`.*\.(?:css|js|bmp|gif|jpe?g|ico|png|tiff?|mid|mp2|mp3|mp4|ppsx|wav|avi|mov|mpeg|ram|m4v|mkv|ogg|ogv|pdf|odc|sas|ps|eps|tex|ppt|pptx|doc|docx|xls|xlsx|names|data|dat|exe|bz2|tar|msi|bin|7z|psd|dmg|iso|epub|dll|cnf|tgz|sha1|ss|scm|py|rkt|r|c|thmx|mso|arff|rtf|jar|csv|java|txt|rm|smil|wmv|swf|wma|zip|rar|gz)$`)
	if urlRe.MatchString(url) && !extRe.MatchString(strings.ToLower(url)) {
		return s.isWordPressWebsite(s.getHostname(url))
	}
	return false
}

func (s *SearchHouseSpider) findHostName(url string) string {
	domainRe := regexp.MustCompile(`https://[^\s:/@]+\.[^\s:/@]+`)
	substr := domainRe.FindAllStringSubmatch(url, -1)
	if len(substr) == 0 || len(substr[0]) == 0 {
		return ""
	}
	return substr[0][0]
}

func (s *SearchHouseSpider) constructProperURLs(urls []string, root string) StringSet {
	var properURLs StringSet
	hostName := s.findHostName(root)
	if hostName == "" {
		return properURLs
	}
	for _, urlStr := range urls {
		var parsedURL string
		if urlStr[0] == '/' {
			parsedURL = hostName + urlStr
		} else {
			parsedURL = strings.TrimSuffix(urlStr, "/")
		}
		if s.urlValid(parsedURL) {
			properURLs.Add(parsedURL)
		}
	}
	return properURLs
}

func (s *SearchHouseSpider) hash(str string) uint64 {
	h := fnv.New64a()
	_, err := h.Write([]byte(str))
	if err != nil {
		log.Fatalln(err)
	}
	return h.Sum64()
}

func (s *SearchHouseSpider) calcWebsiteToRoutineNum(url string) int {
	hostname := s.findHostName(url)
	hash := s.abs(int(s.hash(hostname)))
	return hash % s.numRoutines
}

func (s *SearchHouseSpider) abs(val int) int {
	if val < 0 {
		return -val
	}
	return val
}

func (s *SearchHouseSpider) setSeed(urls []string) {
	for _, urlStr := range urls {
		if !s.pageDownloaded(urlStr) {
			s.frontier.InsertPage(urlStr, 0)
		}
	}
}

func (s *SearchHouseSpider) duplicateExists(fp *common.Fingerprints, wp *common.WebPage) bool {
	fpGlobalSet := fp.GetFingerprintsAsSet()
	fpWebpageSet := wp.Fingerprints.GetFingerprintsAsSet()
	fp.Mu.Lock()
	wp.Fingerprints.Mu.Lock()
	defer wp.Fingerprints.Mu.Unlock()
	defer fp.Mu.Unlock()
	for hash := range fpWebpageSet {
		if pages, exists := fpGlobalSet[hash]; exists {
			for page := range pages {
				if page.Url != wp.Url && wp.Similarity(page) > 0.9 {
					log.Printf("spider - %s has a %f match to %s\n", page.Url, wp.Similarity(page), wp.Url)
					return true
				}
			}
		}
	}
	return false
}

func (s *SearchHouseSpider) validPage(wp *common.WebPage) bool {
	trimmedBody := strings.TrimLeftFunc(wp.Body, unicode.IsSpace)
	return strings.HasPrefix(trimmedBody, "<!DOCTYPE html") || strings.HasPrefix(trimmedBody, "<!doctype html")
}

func (s *SearchHouseSpider) isWordPressWebsite(str string) bool {
	if isWp, exists := s.wordpressSites.Get(str); exists {
		return isWp
	}
	isWp := false
	resp, err := http.Get("https://" + str + "/wp-admin")
	if err == nil && resp != nil {
		content, err := io.ReadAll(resp.Body)
		if err == nil {
			if resp.StatusCode == 403 || (resp.StatusCode == 200 && strings.Contains(strings.ToLower(string(content)), "wordpress")) {
				isWp = true
			}
		}
		resp.Body.Close()
	}
	s.wordpressSites.Add(str, isWp)
	return isWp
}

func (s *SearchHouseSpider) getHostname(u string) string {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		log.Println("spider - Error parsing URL:", err)
		return ""
	}
	return parsedUrl.Host
}
