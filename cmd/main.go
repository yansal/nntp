package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/yansal/nntp"
)

func main() {
	conn, err := nntp.Dial("tcp", "news.gmane.org:119")
	if err != nil {
		log.Fatal(err)
	}

	err = conn.ModeReader()
	if err != nil {
		log.Fatal(err)
	}

	/*
		list, err := conn.List()
		if err != nil {
			log.Fatal(err)
		}
	*/

	group, err := conn.Group("gmane.comp.lang.go.general")
	if err != nil {
		log.Fatal(err)
	}

	startXover := time.Now()
	articles, err := conn.Xover(group)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
	fmt.Println("Xovering took", time.Since(startXover))

	startEncode := time.Now()
	b, err := json.MarshalIndent(articles, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("encoding took", time.Since(startEncode))

	startWrite := time.Now()
	err = ioutil.WriteFile("out.json", b, 0600)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("writing took", time.Since(startWrite))

	f, err := os.Open("out.json")
	if err != nil {
		log.Fatal(err)
	}
	fi, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("file is", fi.Size(), "bytes")
}
