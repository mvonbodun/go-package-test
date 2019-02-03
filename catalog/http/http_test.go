package http

import (
	"bytes"
	"errors"
	"github.com/mvonbodun/go-package-test/catalog"
	"github.com/mvonbodun/go-package-test/catalog/mock"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Global variable for the Handler
var h *Handler

func TestMain(m *testing.M) {
	// Global register the handlers, they can only be run once
	h = NewHandler()

	exitCode := m.Run()

	// Exit
	os.Exit(exitCode)
}

func TestHandler_GetProduct(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	// Mock our Product() call
	ps.ProductFn = func(ctx context.Context, id string) (*catalog.Product, error) {
		if id != "100" {
			t.Fatalf("unexpected id: %v", id)
		}
		return &catalog.Product{
			ID:          "100",
			ProductCode: "abcdef",
			ShortDesc:   "This is the short description.",
			LongDesc:    "This is the long description.",
		}, nil
	}

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/product/100", nil)
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if !ps.ProductInvoked {
		t.Fatal("expected Product() to be invoked.")
	}
}

func TestHandler_GetProductNotFound(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	// Mock our Product() call
	ps.ProductFn = func(ctx context.Context, id string) (*catalog.Product, error) {
		return &catalog.Product{}, errors.New("not found")
	}

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/product/99", nil)
	h.Router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Fatal("Should not have received http status 404. Record not found.")
	}

	// Validate mock.
	if !ps.ProductInvoked {
		t.Fatal("expected Product() to be invoked.")
	}
}

func TestHandler_GetProducts(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.ProductsFn = func(ctx context.Context) ([]*catalog.Product, error) {

		products := []*catalog.Product{
			{
				ID:          "100",
				ProductCode: "abcdef",
				ShortDesc:   "This is the short description.",
				LongDesc:    "This is the long description.",
			},
			{
				ID:          "200",
				ProductCode: "lmnopq",
				ShortDesc:   "This is the short description 2.",
				LongDesc:    "This is the long description 2.",
			},
		}

		return products, nil
	}

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/products", nil)
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if !ps.ProductsInvoked {
		t.Fatal("expect Products() to be invoked.")
	}

}

func TestHandler_GetProductsNotFound(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.ProductsFn = func(ctx context.Context) ([]*catalog.Product, error) {

		products := []*catalog.Product{
			{},
		}

		return products, errors.New("no products found")
	}

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/products", nil)
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if !ps.ProductsInvoked {
		t.Fatal("expect Products() to be invoked.")
	}
	if w.Code != http.StatusNotFound {
		t.Fatal("expected 404 status code")
	}

}

func TestHandler_AddProduct(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.CreateProductFn = func(ctx context.Context, product *catalog.Product) error {
		return nil
	}

	payload := []byte(`{  "productCode": "prod15", "shortDesc": "Short desc for prod 15", "longDesc": "Long desc for prod 15" }`)

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if !ps.CreateProductInvoked {
		t.Fatal("expect CreateProduct() to be invoked.")
	}
	if w.Code != http.StatusCreated {
		t.Fatal("expected 201 status code")
	}

}

func TestHandler_AddProductBadJson(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.CreateProductFn = func(ctx context.Context, product *catalog.Product) error {
		return nil
	}

	payload := []byte(`<junk>`)

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if w.Code != http.StatusBadRequest {
		t.Fatal("expected 400 status code")
	}

}

func TestHandler_AddProductFailedInsert(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.CreateProductFn = func(ctx context.Context, product *catalog.Product) error {
		return errors.New("error inserting product")
	}

	payload := []byte(`{  "productCode": "prod15", "shortDesc": "Short desc for prod 15", "longDesc": "Long desc for prod 15" }`)

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/product", bytes.NewBuffer(payload))
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if w.Code != http.StatusBadRequest {
		t.Fatal("expected 400 status code")
	}

}

func TestHandler_DeleteProduct(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.DeleteProductFn = func(ctx context.Context, id string) error {
		return nil
	}

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/product/100", nil)
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if !ps.DeleteProductInvoked {
		t.Fatal("expect DeleteProduct() to be invoked.")
	}
	if w.Code != http.StatusOK {
		t.Fatal("expected 200 status code")
	}

}

func TestHandler_DeleteProductNotFound(t *testing.T) {
	// Inject our mock into our handler.
	var ps mock.ProductService
	h.ProductService = &ps

	ps.DeleteProductFn = func(ctx context.Context, id string) error {
		return errors.New("product not found")
	}

	// Invoke the handler.
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("DELETE", "/product/100", nil)
	h.Router.ServeHTTP(w, r)

	// Validate mock.
	if !ps.DeleteProductInvoked {
		t.Fatal("expect DeleteProduct() to be invoked.")
	}
	if w.Code != http.StatusNotFound {
		t.Fatal("expected 404 status code")
	}

}
