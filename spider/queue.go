package spider

import "sync"

type StringQueue struct {
	q  []string
	mu sync.Mutex
}

func (sq *StringQueue) Insert(str string) {
	sq.mu.Lock()
	sq.q = append(sq.q, str)
	sq.mu.Unlock()
}

func (sq *StringQueue) Pop() string {
	sq.mu.Lock()
	str := sq.q[0]
	sq.q = sq.q[1:]
	sq.mu.Unlock()
	return str
}

func (sq *StringQueue) Len() int {
	sq.mu.Lock()
	length := len(sq.q)
	sq.mu.Unlock()
	return length
}
