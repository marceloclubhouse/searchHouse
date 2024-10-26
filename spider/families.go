package spider

import (
	"fmt"
	"strings"
)

type UselessFamilies struct {
	db    map[string]struct{}
	queue []string
	n     int
}

func NewUselessFamilies(n int) *UselessFamilies {
	return &UselessFamilies{
		db:    make(map[string]struct{}),
		queue: make([]string, 0),
		n:     n,
	}
}

func (uf *UselessFamilies) Insert(url string) {
	uf.queue = append(uf.queue, url)
	if len(uf.queue) > uf.n {
		if uf.checkFamily() {
			family := uf.generateFamilyRule()
			if _, exists := uf.db[family]; !exists {
				uf.addFamily(family)
			}
		}
		uf.queue = uf.queue[1:]
	}
}

func (uf *UselessFamilies) Useless(url string) bool {
	for family := range uf.db {
		if strings.Contains(url, family) {
			fmt.Printf("USELESS FAMILY:\tPage %s determined to come from a useless family. Excluding from frontier.\n", url)
			return true
		}
	}
	return false
}

func (uf *UselessFamilies) checkFamily() bool {
	length := len(uf.queue[0])
	for _, url := range uf.queue[1:] {
		if len(url) != length {
			return false
		}
	}
	return true
}

func (uf *UselessFamilies) generateFamilyRule() string {
	convergedStr := uf.queue[0]
	for _, url := range uf.queue[1:] {
		convergedStr = uf.strUnion(convergedStr, url)
	}
	return convergedStr
}

func (uf *UselessFamilies) addFamily(family string) {
	if len(family) > 0 {
		uf.db[family] = struct{}{}
	}
}

func (uf *UselessFamilies) strUnion(str1, str2 string) string {
	overlap := make([]string, 0)
	unionStr := ""
	i := 0
	for i < len(str1) {
		if str1[i] == str2[i] {
			unionStr += string(str1[i])
		} else {
			overlap = append(overlap, unionStr)
			unionStr = ""
		}
		i++
	}
	if len(overlap) == 0 {
		return str1
	}
	maxLength := 0
	maxIndex := 0
	for i, str := range overlap {
		if len(str) > maxLength {
			maxLength = len(str)
			maxIndex = i
		}
	}
	return overlap[maxIndex]
}
