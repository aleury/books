package books_test

import (
	"books"
	"cmp"
	"net"
	"slices"
	"testing"
)

var (
	ABC = books.Book{
		ID:     "abc",
		Title:  "In the Company of Cheeful Ladies",
		Author: "Alexander McCall Smith",
		Copies: 1,
	}
	XYZ = books.Book{
		ID:     "xyz",
		Title:  "White Heat",
		Author: "Dominic Sandbrook",
		Copies: 2,
	}
)

func TestString_FormatsBookInfoAsString(t *testing.T) {
	t.Parallel()
	book := books.Book{
		Title:  "Sea Room",
		Author: "Adam Nicolson",
		Copies: 2,
	}
	want := "Sea Room by Adam Nicolson (copies: 2)"
	got := book.String()
	if want != got {
		t.Fatalf("want %q, got %q", want, got)
	}
}

func TestSetCopies_SetsNumberOfCopiesToGivenValue(t *testing.T) {
	t.Parallel()
	book := books.Book{
		Copies: 5,
	}
	err := book.SetCopies(12)
	if err != nil {
		t.Fatal(err)
	}
	if book.Copies != 12 {
		t.Fatalf("want 12 copies, got %d", book.Copies)
	}
}

func TestNewCatalog_ReturnsAnEmptyCatalog(t *testing.T) {
	t.Parallel()
	catalog := books.NewCatalog()
	got := catalog.GetAllBooks()
	if len(got) > 0 {
		t.Fatalf("want empty catalog, got %v", got)
	}
}

func TestSetCopies_ReturnsErrorIfCopiesNegative(t *testing.T) {
	t.Parallel()
	book := books.Book{}
	err := book.SetCopies(-1)
	if err == nil {
		t.Fatal("expected error for negative copies, got nil")
	}
}

func TestSetCopies_OnCatalogModifiesSpecifiedBook(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	book, ok := catalog.GetBook("abc")
	if !ok {
		t.Fatal("book not found")
	}
	if book.Copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", book.Copies)
	}
	err := catalog.SetCopies("abc", 2)
	if err != nil {
		t.Fatal(err)
	}
	book, ok = catalog.GetBook("abc")
	if !ok {
		t.Fatal("book not found")
	}
	if book.Copies != 2 {
		t.Fatalf("want 2 copies after change, got %d", book.Copies)
	}
}

func TestOpenCatalog_ReadsSameDataWrittenBySync(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	catalog.Path = t.TempDir() + "/catalog"
	err := catalog.Sync()
	if err != nil {
		t.Fatal(err)
	}
	newCatalog, err := books.OpenCatalog(catalog.Path)
	if err != nil {
		t.Fatal(err)
	}
	got := newCatalog.GetAllBooks()
	assertTestBooks(t, got)
}

func TestGetAllBooks_ReturnsAllBooks(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	got := catalog.GetAllBooks()
	assertTestBooks(t, got)
}

func TestGetBook_FindsBookInCatalogByID(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	got, ok := catalog.GetBook("abc")
	if !ok {
		t.Fatalf("book not found")
	}
	if got != ABC {
		t.Fatalf("want %#v, got %#v", ABC, got)
	}
}

func TestGetBook_ReturnsFalseWhenBookNotFound(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	_, ok := catalog.GetBook("nonexistent ID")
	if ok {
		t.Fatal("want false for nonexistent ID, got true")
	}
}

func TestAddBook_AddsGivenBookToCatalog(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	_, ok := catalog.GetBook("123")
	if ok {
		t.Fatal("book already present")
	}
	err := catalog.AddBook(books.Book{
		ID:     "123",
		Title:  "The Prize of all the Oceans",
		Author: "Glyn Williams",
		Copies: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, ok = catalog.GetBook("123")
	if !ok {
		t.Fatal("added book not found")
	}
}

func TestAddBook_ReturnsErrorIfIDExists(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	if _, ok := catalog.GetBook("abc"); !ok {
		t.Fatal("book not present")
	}
	err := catalog.AddBook(ABC)
	if err == nil {
		t.Fatal("want error for duplicate id, got nil")
	}
}

func TestSetCopies_IsRaceFree(t *testing.T) {
	t.Parallel()
	catalog := getTestCatalog()
	go func() {
		for range 100 {
			err := catalog.SetCopies("abc", 0)
			if err != nil {
				panic(err)
			}
		}
	}()
	for range 100 {
		_, err := catalog.GetCopies("abc")
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetAllBooks_OnClientListsAllBooks(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	bookList, err := client.GetAllBooks()
	if err != nil {
		t.Fatal(err)
	}
	assertTestBooks(t, bookList)
}

func TestGetBook_OnClientFindsBookByID(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	got, err := client.GetBook("abc")
	if err != nil {
		t.Fatal(err)
	}
	if got != ABC {
		t.Fatalf("want %#v, got %#v", ABC, got)
	}
}

func TestGetBook_OnClientReturnsErrorWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.GetBook("bogus")
	if err == nil {
		t.Fatal("want error when book not found, got nil")
	}
}

func TestGetCopies_OnClientReturnsCopiesForBook(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	got, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("want 1, got %d", got)
	}
}

func TestGetCopies_OnClientReturnsErrorWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.GetCopies("bogus")
	if err == nil {
		t.Fatal("want error when book not found, got nil")
	}
}

func TestAddCopies_OnClientCorrectlyUpdatesStockLevel(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	copies, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", copies)
	}
	stock, err := client.AddCopies("abc", 2)
	if err != nil {
		t.Fatal(err)
	}
	if stock != 3 {
		t.Fatalf("want 3 after change, got %d", stock)
	}
}

func TestAddCopies_OnClientReturnsErrorWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.AddCopies("bogus", 2)
	if err == nil {
		t.Fatal("want error when book not found, got nil")
	}
}

func TestSubCopies_OnClientCorrectlyUpdatesStockLevel(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	copies, err := client.GetCopies("abc")
	if err != nil {
		t.Fatal(err)
	}
	if copies != 1 {
		t.Fatalf("want 1 copy before change, got %d", copies)
	}
	stock, err := client.SubCopies("abc", 1)
	if err != nil {
		t.Fatal(err)
	}
	if stock != 0 {
		t.Fatalf("want 0 after change, got %d", stock)
	}
}

func TestSubCopies_OnClientReturnsErrorWhenNotEnoughStock(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.SubCopies("abc", 2)
	if err == nil {
		t.Fatal("want error when not enough stock, got nil")
	}
}

func TestSubCopies_OnClientReturnsErrorWhenBookNotFound(t *testing.T) {
	t.Parallel()
	client := getTestClient(t)
	_, err := client.SubCopies("bogus", 1)
	if err == nil {
		t.Fatal("want error when book not found, got nil")
	}
}

func getTestCatalog() *books.Catalog {
	catalog := books.NewCatalog()
	err := catalog.AddBook(ABC)
	if err != nil {
		panic(err)
	}
	err = catalog.AddBook(XYZ)
	if err != nil {
		panic(err)
	}
	return catalog
}

func assertTestBooks(t *testing.T, got []books.Book) {
	t.Helper()
	want := []books.Book{ABC, XYZ}
	slices.SortFunc(got, func(a, b books.Book) int {
		return cmp.Compare(a.Author, b.Author)
	})
	if !slices.Equal(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}

func randomLocalAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().String()
}

func getTestClient(t *testing.T) *books.Client {
	t.Helper()
	addr := randomLocalAddr(t)
	catalog := getTestCatalog()
	catalog.Path = t.TempDir() + "/catalog"
	go func() {
		err := books.ListenAndServe(addr, catalog)
		if err != nil {
			panic(err)
		}
	}()
	return books.NewClient(addr)
}
