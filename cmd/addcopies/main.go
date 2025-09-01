package main

import (
	"books"
	"fmt"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: addcopies <BOOK ID> <HOW MANY>")
		return
	}
	id := os.Args[1]
	copies, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Printf("invalid number of copies: %v\n", err)
		return
	}
	client := books.NewClient("localhost:3000")
	stock, err := client.AddCopies(id, copies)
	if err != nil {
		fmt.Printf("error adding copies: %v\n", err)
		return
	}
	fmt.Printf("%d copies in stock\n", stock)
}
