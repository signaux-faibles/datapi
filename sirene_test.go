package main

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestIds(t *testing.T) {
	t.Log("ids returns valid arrays")
	i := ids(0, 50)
	if len(i) != 50 {
		t.Fatal()
	}
	for k, v := range i {
		if v != k+1 {
			fmt.Println(len(i), k, v)
			t.Fatal()
		}
	}
	i = ids(5, 50)
	if len(i) != 50 {
		t.Fatal()
	}
	for k, v := range i {
		if v != k+251 {
			fmt.Println(len(i), k, v)
			t.Fatal()
		}
	}
}

func TestGetInsertSireneUL(t *testing.T) {
	t.Log("getInsertSireneUL returns expected sql statement")
	sql := getInsertSireneUL(3)
	sum := md5.Sum([]byte(sql[1]))
	if fmt.Sprintf("%x", sum) != "ff7c3021e1251513f8d3d1a58ab50c91" {
		t.Fatal()
	}
}
