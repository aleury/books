package main

import (
	"books"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: getcopies <BOOK ID>")
		return
	}
	client := books.NewClient("localhost:3000")
	id := os.Args[1]
	copies, err := client.GetCopies(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%d copies in stock\n", copies)
}
