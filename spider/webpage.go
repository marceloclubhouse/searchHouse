package spider

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"
)

type WebPage struct {
	Time         int64  `json:"time"`
	Url          string `json:"url"`
	Response     string `json:"response"`
	Body         string `json:"body"`
	fingerprints *Fingerprints
}

func NewWebPage(time int64, url string, response string, body string) *WebPage {
	wp := &WebPage{
		Time:         time,
		Url:          url,
		Response:     response,
		Body:         body,
		fingerprints: NewFingerprints(3, 1000),
	}
	wp.fingerprints.InsertFingerprintsUsingWebpage(wp)
	return wp
}

func (wp *WebPage) Serialize() []byte {
	// Serialize WebPage object as JSON byte array
	b, err := json.Marshal(wp)
	if err != nil {
		log.Fatalln(err)
	}
	return b
}

func (wp *WebPage) FindAllAnchorHREFs(maxNumHREF int) []string {
	// Find all links within HTML markup
	// (<a href="...">) -> ["..."]
	var hrefs []string
	re := regexp.MustCompile(`href=['"]?([^'" >]+)`)
	matches := re.FindAllStringSubmatch(wp.Body, maxNumHREF)
	for _, element := range matches {
		hrefs = append(hrefs, element[1])
	}
	return hrefs
}

func (wp *WebPage) findAllTags(tags []string) []string {
	// Extract the specified tags out of HTML
	// markup and return the content of each
	// tag as a list of strings
	var content []string
	for _, tag := range tags {
		re := regexp.MustCompile(fmt.Sprintf(`<%s.*?>(.*)</%s>`, tag, tag))
		matches := re.FindAllStringSubmatch(wp.Body, -1)
		for _, element := range matches {
			content = append(content, element[1])
		}
	}
	return content
}

func (wp *WebPage) removeAllMarkup(str string) string {
	// Given some text, remove all HTML brackets
	// from it (e.g. "this <span>is</span> text"
	// becomes "this is text")
	str = html.UnescapeString(str)
	re := regexp.MustCompile(`<.*?>|</.*>`)
	if re.MatchString(str) {
		return re.ReplaceAllString(str, " ")
	} else {
		return str
	}
}

func (wp *WebPage) removeAllPunctuation(str string) string {
	re := regexp.MustCompile(`[^a-zA-Z0-9-:\s'’/–-]`)
	if re.MatchString(str) {
		return re.ReplaceAllString(str, " ")
	} else {
		return str
	}
}

func (wp *WebPage) extractText() string {
	// Extract all the text from this WebPage object
	var str string
	tags := wp.findAllTags([]string{"p"})
	for _, content := range tags {
		str = str + strings.ToLower(wp.removeAllPunctuation(wp.removeAllMarkup(content)))
	}
	return str
}

func (wp *WebPage) Similarity(webPage *WebPage) float64 {
	intersection := 0
	left := wp.fingerprints.GetFingerprintsAsSet()
	right := webPage.fingerprints.GetFingerprintsAsSet()
	for hash := range left {
		if _, exists := right[hash]; exists {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	return float64(intersection) / float64(union)
}
