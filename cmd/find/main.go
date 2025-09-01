package main

import (
	"books"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: find <BOOK ID>")
		return
	}
	client := books.NewClient("localhost:3000")
	book, err := client.GetBook(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(book)
}
