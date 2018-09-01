package http

import (
	"github.com/mvonbodun/go-package-test/catalog"
	"net/http"
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"encoding/json"
	"github.com/gorilla/handlers"
	"os"
	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/net/context"
)

var (
	h1 *Handler
)

type Handler struct {
	ProductService catalog.ProductService
	Handler *Handler
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	h := &Handler{}
	return h
}

//// ServeHTTP handles the requests.
//func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	h.registerHandlers()
//}

// StartWebServer starts the web server
func (h *Handler) ListenAndServe() {
	h.registerHandlers()
	h1 = h.Handler
	log.Info(http.ListenAndServe(":8080", &ochttp.Handler{}))
}

// registerHandlers registers the handlers used to respond to requests.
func (h *Handler) registerHandlers() {
	// Use gorilla/mux for rich routing
	r := mux.NewRouter()
	//  All API calls leverage application/json
	s := r.Headers("Accept", "application/json").Subrouter()

	s.Methods("GET").Path("/product/{id:[0-9]+}").
		HandlerFunc(GetProduct)

	s.Methods("GET").Path("/product").
		HandlerFunc(GetProducts)

	s.Methods("POST").Path("/product").
		HandlerFunc(AddProduct)

	s.Methods("DELETE").Path("/product/{id:[0-9]+}").
		HandlerFunc(DeleteProduct)

	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, r))
}

// GetProduct retrieves a single product from the database.
func GetProduct(w http.ResponseWriter, r *http.Request) {
	// Get the variables from the request
	vars := mux.Vars(r)
	productId := vars["id"]
	log.Info("id=%v\n", productId)
	product, err := h1.getProduct(r.Context(),productId)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Error retrieving product: %v\n", err)
		log.Error(err)
	} else {
		p, err := json.Marshal(product)
		if err != nil {
			log.Infof("error: %v\n", err)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, string(p))
	}
}

func (h *Handler) getProduct(ctx context.Context, productId string) (*catalog.Product, error) {
	if len(productId) == 0 {
		log.Fatal("Error productId not passed in")
	}
	log.Info("Inside getProduct.")
	// Get the produce from the database
	product, err := h.ProductService.Product(ctx, productId)
	return product, err
}

// GetProducts retrieves all of the products from the database.
func GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h1.getProducts(r.Context())
	if err != nil {
		fmt.Fprintf(w, "An error occured: %v\n", err)
	} else {
		// Discard the index, get the value
		//for _, v := range products {
		//	fmt.Fprintf(w, "productCode: %v\n", v.ProductCode)
		//}
		p, err := json.Marshal(products)
		if err != nil {
			log.Infof("error: %v\n", err)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, string(p))
	}
}

func (h *Handler) getProducts(ctx context.Context) ([]*catalog.Product, error) {
	products, err := h.ProductService.Products(ctx)
	return products, err
}

// AddProduct adds a single product to the database.
func AddProduct(w http.ResponseWriter, r *http.Request) {
	product := &catalog.Product{}
	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		log.Fatal("Error decoding: %v", err)
	}
	log.Info("The body that was posted: %v\n", product.ProductCode)
	// Add the catalog to the database
	err := h1.ProductService.CreateProduct(r.Context(), product)
	if err != nil {
		log.Fatal("Error adding product: %v\n", err)
	} else {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Successfully added product to DB.\n")
	}
}

func (h *Handler) addProduct(ctx context.Context, product catalog.Product) error {
	err := h.ProductService.CreateProduct(ctx, &product)
	return err
}

// DeleteProduct deletes a product from the database.
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Get the variables from the request
	vars := mux.Vars(r)
	productId := vars["id"]
	log.Info("id=%v\n", productId)
	err := h1.ProductService.DeleteProduct(r.Context(), productId)
	if err != nil {
		fmt.Fprintf(w, "Error when deleting product with id: %v\n", productId)
	} else {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "Product with id: %v was deleted.\n", productId)
	}
}

func (h *Handler) deleteProduct(ctx context.Context, id string) error {
	err := h.ProductService.DeleteProduct(ctx, id)
	return err
}