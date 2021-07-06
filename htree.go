package main

import (
	"modernc.org/b"
)

type htree struct {
	tree *b.Tree
}

func (t *htree) contains(hash []byte) (int, bool) {
	if t.tree == nil {
		return 0, false
	}
	id, ok := t.tree.Get(hash)
	if !ok {
		return 0, ok
	}
	return id.(int), ok
}
