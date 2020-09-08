package main

import (
	"bytes"

	"modernc.org/b"
)

type htree struct {
	tree *b.Tree
}

func (t *htree) insert(hash []byte, id int) {
	if t.tree == nil {
		t.tree = b.TreeNew(hashCmp)
	}
	t.tree.Set(hash, id)
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

func hashCmp(a, b interface{}) int {
	return bytes.Compare((a.([]byte)), (b.([]byte)))
}
