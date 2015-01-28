package main

import (
	"sync"
	"testing"
)

func TestIdGenerator(t *testing.T) {
	const (
		l = 5
		n = 10000
	)

	for j := 0; j < l; j++ {
		idGen := newIdGenerator(l)

		m := map[string]bool{}
		for i := 0; i < n; i++ {
			id, err := idGen.newId()
			if err != nil {
				t.Fatal(err)
			} else if m[id] {
				t.Error(id)
			}
			m[id] = true
		}
	}
}

func TestIdGeneratorConcurrent(t *testing.T) {
	const (
		p = 10
		n = 5000
	)

	idGen := newIdGenerator(0)

	glbl := map[string]bool{}
	var lock sync.Mutex
	errCh := make(chan int, p)
	endCh := make(chan struct{}, p)

	for j := 0; j < p; j++ {
		go func(self int) {
			defer func() { endCh <- struct{}{} }()

			// まずローカルで ID 生成する。
			buff := map[string]bool{}
			for i := 0; i < n; i++ {
				id, err := idGen.newId()
				if err != nil {
					t.Error(err)
					errCh <- self
					return
				} else if buff[id] {
					errCh <- self
					return
				}
				buff[id] = true
			}

			// グローバルに ID が重複していないか検査。
			lock.Lock()
			defer lock.Unlock()
			for id := range buff {
				if glbl[id] {
					errCh <- self
					return
				}
				glbl[id] = true
			}
		}(j)
	}

	for i := 0; i < p; i++ {
		<-endCh
	}
	for ok := false; ok; {
		select {
		case pid := <-errCh:
			t.Error(pid)
		default:
			ok = true
		}
	}
	if len(glbl) != p*n {
		t.Error(len(glbl), p*n)
	}
}
