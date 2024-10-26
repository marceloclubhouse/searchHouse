package common

import (
	"hash/fnv"
	"strings"
	"sync"
)

// Fingerprints struct is a hashmap that maps integers
// to a list of WebPage pointers. Each integer is a hashed
// fingerprint.

type Fingerprints struct {
	Mu      sync.Mutex
	n       int
	fpSet   map[uint32]map[*WebPage]bool
	maxSize int
}

func NewFingerprints(n int, maxSize int) *Fingerprints {
	return &Fingerprints{n: n, fpSet: make(map[uint32]map[*WebPage]bool), maxSize: maxSize}
}

func (fp *Fingerprints) nGram(text string) []string {
	nGrams := make([]string, 0)
	words := strings.Fields(strings.ToLower(text))
	end := len(words) - fp.n + 1
	for i := 0; i < end; i++ {
		nGram := strings.Join(words[i:i+fp.n], " ")
		nGrams = append(nGrams, nGram)
	}
	return nGrams
}

func (fp *Fingerprints) nGramsToHashes(nGrams []string) []uint32 {
	hashes := make([]uint32, len(nGrams))
	for i, nGram := range nGrams {
		hashes[i] = fnvHash(nGram)
	}
	return hashes
}

func (fp *Fingerprints) InsertFingerprintsUsingWebpage(wp *WebPage) {
	nGrams := fp.nGram(wp.Body)
	hashes := fp.nGramsToHashes(nGrams)
	fp.Mu.Lock()
	for _, h := range hashes {
		if h%uint32(fp.n) == 0 {
			fp.maintainSizeLimit()
			if _, exists := fp.fpSet[h]; exists {
				fp.fpSet[h][wp] = true
			} else {
				fp.fpSet[h] = make(map[*WebPage]bool)
				fp.fpSet[h][wp] = true
			}
		}
	}
	fp.Mu.Unlock()
}

func (fp *Fingerprints) GetFingerprintsAsSet() map[uint32]map[*WebPage]bool {
	return fp.fpSet
}

func fnvHash(s string) uint32 {
	hash := fnv.New32()
	_, _ = hash.Write([]byte(s))
	return hash.Sum32()
}

func (fp *Fingerprints) maintainSizeLimit() {
	// Reset if the size is too large
	// TODO: Implement better eviction policy
	if len(fp.fpSet) > fp.maxSize {
		fp.fpSet = make(map[uint32]map[*WebPage]bool)
	}
}
