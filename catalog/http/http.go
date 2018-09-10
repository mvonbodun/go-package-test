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
	"net/http/httputil"
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

// StartWebServer starts the web server
func (h *Handler) ListenAndServe() {
	h.registerHandlers()
	log.Info(http.ListenAndServe(":8080", &ochttp.Handler{}))
}

// registerHandlers registers the handlers used to respond to requests.
func (h *Handler) registerHandlers() {
	// Use gorilla/mux for rich routing
	r := mux.NewRouter()
	//  All API calls leverage application/json
	s := r.Headers("Accept", "application/json").Subrouter()

	s.Methods("GET").Path("/product/{id:[0-9]+}").
		HandlerFunc(h.GetProduct)

	s.Methods("GET").Path("/product").
		HandlerFunc(h.GetProducts)

	s.Methods("POST").Path("/product").
		HandlerFunc(h.AddProduct)

	s.Methods("DELETE").Path("/product/{id:[0-9]+}").
		HandlerFunc(h.DeleteProduct)

	http.Handle("/", handlers.CombinedLoggingHandler(os.Stdout, r))
}

// GetProduct retrieves a single product from the database.
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Get the variables from the request
	vars := mux.Vars(r)
	productId := vars["id"]
	if len(productId) == 0 {
		log.WithField("httpRequest", r).
			Warning("Error productId not passed in.")
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w,"productId not passed in.")
	} else {
		product, err := h.ProductService.Product(r.Context(), productId)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintf(w, "Warning: no product found: %v", err)
			log.WithField("httpRequest", r).
				Warningf("No product was found: %v", err)
		} else {
			p, err := json.Marshal(product)
			if err != nil {
				log.WithField("httpRequest", r).
					Errorf("Error marshalling product: %v", err)
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintf(w, string(p))
		}
	}
}

// GetProducts retrieves all of the products from the database.
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	// Print the request header
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.WithField("httpRequest", r).
			Errorf("Failed to get request dump: %v", err)
	} else {
		log.WithField("httpRequest", r).
			Debugf("Request header dump: %v", string(requestDump))
	}
	products, err := h.ProductService.Products(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "An error occured retrieving products: %v", err)
	} else {
		p, err := json.Marshal(products)
		if err != nil {
			log.WithField("httpRequest", r).
				Errorf("error marshalling: %v", err)
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, string(p))
	}
}

// AddProduct adds a single product to the database.
func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	product := &catalog.Product{}
	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		log.WithField("httpRequest", r).
			Errorf("Error decoding during AddProduct: %v", err)
	}
	log.Debugf("The body that was posted for ProductCode: %v", product.ProductCode)
	// Add the catalog to the database
	err := h.ProductService.CreateProduct(r.Context(), product)
	if err != nil {
		log.WithField("httpRequest", r).
			Errorf("Error adding product: %v", err)
	} else {
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, "Successfully added product to DB.")
	}
}

// DeleteProduct deletes a product from the database.
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Get the variables from the request
	vars := mux.Vars(r)
	productId := vars["id"]
	if len(productId) == 0 {
		log.WithField("httpRequest", r).
			Warning("Error productId not passed in.")
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w,"productId not passed in.")
	} else {
		log.Debugf("From the request productId=%v", productId)
		err := h.ProductService.DeleteProduct(r.Context(), productId)
		if err != nil {
			fmt.Fprintf(w, "Error when deleting product with id: %v", productId)
			log.WithField("httpRequest", r).
				Warningf("No product was found during delete: %v", err)
		} else {
			w.WriteHeader(http.StatusNoContent)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			fmt.Fprintf(w, "Product with id: %v was deleted.", productId)
		}
	}
}
