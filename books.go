package books

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"
)

var ErrNotEnoughStock = errors.New("not enough stock")

type Book struct {
	ID     string
	Title  string
	Author string
	Copies int
}

func (book Book) String() string {
	return fmt.Sprintf("%v by %v (copies: %v)", book.Title, book.Author, book.Copies)
}

func (book *Book) SetCopies(copies int) error {
	if copies < 0 {
		return fmt.Errorf("negative number of copies: %d", copies)
	}
	book.Copies = copies
	return nil
}

type Catalog struct {
	mu   *sync.RWMutex
	data map[string]Book
	Path string
}

func NewCatalog() *Catalog {
	return &Catalog{
		mu:   &sync.RWMutex{},
		data: map[string]Book{},
	}
}

func OpenCatalog(path string) (*Catalog, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	catalog := NewCatalog()
	err = json.NewDecoder(file).Decode(&catalog.data)
	if err != nil {
		return nil, err
	}
	catalog.Path = path
	return catalog, nil
}

func (catalog *Catalog) GetAllBooks() []Book {
	catalog.mu.RLock()
	defer catalog.mu.RUnlock()
	return slices.Collect(maps.Values(catalog.data))
}

func (catalog *Catalog) GetBook(id string) (Book, bool) {
	catalog.mu.RLock()
	defer catalog.mu.RUnlock()
	book, ok := catalog.data[id]
	return book, ok
}

func (catalog *Catalog) AddBook(book Book) error {
	catalog.mu.Lock()
	defer catalog.mu.Unlock()
	if _, ok := catalog.data[book.ID]; ok {
		return fmt.Errorf("id %q already exists", book.ID)
	}
	catalog.data[book.ID] = book
	return nil
}

func (catalog *Catalog) AddCopies(id string, copies int) (int, error) {
	catalog.mu.Lock()
	defer catalog.mu.Unlock()
	book, ok := catalog.data[id]
	if !ok {
		return 0, fmt.Errorf("id %q not found", id)
	}
	book.Copies += copies
	catalog.data[book.ID] = book
	return book.Copies, nil
}

func (catalog *Catalog) SubCopies(id string, copies int) (int, error) {
	catalog.mu.Lock()
	defer catalog.mu.Unlock()
	book, ok := catalog.data[id]
	if !ok {
		return 0, fmt.Errorf("id %q not found", id)
	}
	if book.Copies < copies {
		return 0, fmt.Errorf("%w: %d", ErrNotEnoughStock, book.Copies)
	}
	book.Copies -= copies
	catalog.data[book.ID] = book
	return book.Copies, nil
}

func (catalog *Catalog) GetCopies(id string) (int, error) {
	catalog.mu.RLock()
	defer catalog.mu.RUnlock()
	book, ok := catalog.data[id]
	if !ok {
		return 0, fmt.Errorf("id %q not found", id)
	}
	return book.Copies, nil
}

func (catalog *Catalog) SetCopies(id string, copies int) error {
	catalog.mu.Lock()
	defer catalog.mu.Unlock()
	book, ok := catalog.data[id]
	if !ok {
		return fmt.Errorf("id %q not found", id)
	}
	err := book.SetCopies(copies)
	if err != nil {
		return err
	}
	catalog.data[book.ID] = book
	return nil
}

func (catalog *Catalog) Sync() error {
	catalog.mu.RLock()
	defer catalog.mu.RUnlock()
	file, err := os.Create(catalog.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = json.NewEncoder(file).Encode(catalog.data)
	if err != nil {
		return err
	}
	return nil
}
