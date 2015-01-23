package main

import (
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
