package main

import (
	"log"
	"github.com/spf13/cobra/doc"
	"github.com/rajdeepvala/subextract/cmd"
)

func main() {
	header := &doc.GenManHeader{
		Title:   "SUBEXTRACT",
		Section: "1",
	}
	err := doc.GenManTree(cmd.GetRoot(), header, "./docs")
	if err != nil {
		log.Fatal(err)
	}
}
