package main

import (
	"bytes"
	"crypto/md5"
	"fmt"

	"github.com/gin-gonic/gin"
	"modernc.org/b"
)

type htree struct {
	tree *b.Tree
}

func (t *htree) insert(hash []byte) {
	if t.tree == nil {
		t.tree = b.TreeNew(hashCmp)
	}
	t.tree.Set(hash, struct{}{})
}

func (t *htree) contains(hash []byte) bool {
	_, ok := t.tree.Get(hash)
	return ok
}

func test(c *gin.Context) {
	t := b.TreeNew(hashCmp)
	a := md5.Sum([]byte("blcblcblcb"))
	b := md5.Sum([]byte("blcblcblc."))
	d := md5.Sum([]byte("blcblcblcpeé"))
	t.Set(a[:], struct{}{})
	t.Set(b[:], struct{}{})

	i, err := t.Get(a[:])
	fmt.Println(i, err)
	j, err := t.Get(b[:])
	fmt.Println(j, err)
	k, err := t.Get(d[:])
	fmt.Println(k, err)

}

func hashCmp(a, b interface{}) int {
	return bytes.Compare((a.([]byte)), (b.([]byte)))
}
