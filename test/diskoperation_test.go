package test

import (
	"github.com/Pigeon377/Dandelion/skiplist"
	"log"
	"os"
	"testing"
)

func TestWriteDBFile(t *testing.T) {
	skipList := skiplist.NewSkipList(17)
	skipList.Put(778, []byte("sda"))
	skipList.Put(7738, []byte("1"))
	skipList.Put(72178, []byte("true"))
	skipList.Put(71378, []byte("114"))
	skipList.Put(72378, []byte("114"))
	skipList.Put(73478, []byte("???"))
	skipList.Put(77568, []byte("&&&"))
	f, err := os.OpenFile("test.txt", os.O_WRONLY|os.O_CREATE, 0666)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
	if err != nil {
		log.Fatalln(err)
	}
}
