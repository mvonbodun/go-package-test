package http

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mvonbodun/go-package-test/catalog"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"strings"
)

type Handler struct {
	ProductService catalog.ProductService
	Handler        *Handler
	Router         *mux.Router
}

// NewHandler creates a new Handler.
func NewHandler() *Handler {
	h := &Handler{}
	h.Router = h.registerHandlers()
	return h
}

// registerHandlers registers the handlers used to respond to requests.
func (h *Handler) registerHandlers() *mux.Router {
	// Use gorilla/mux for rich routing
	r := mux.NewRouter()
	//r.PathPrefix("/product")
	//  All API calls leverage application/json
	s := r.Headers("Accept", "application/json").Subrouter()

	read := s.Methods("GET").Handler(negroni.New(negroni.HandlerFunc(readMiddleware)))
	write := s.Methods("POST", "PUT", "DELETE").Handler(negroni.New(negroni.HandlerFunc(writeMiddleware)))

	read.Path("/product/{id:[0-9]+}").
		HandlerFunc(h.GetProduct)

	read.Path("/products").
		HandlerFunc(h.GetProducts)

	write.Path("/product").
		HandlerFunc(h.AddProduct)

	write.Path("/product").
		HandlerFunc(h.UpdateProduct)

	write.Path("/product/{id:[0-9]+}").
		HandlerFunc(h.DeleteProduct)

	http.Handle("/", handlers.CompressHandler(handlers.CombinedLoggingHandler(os.Stdout, r)))

	return s
}

//func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
//	// Do nothing
//	log.Debug("http: Entered ServerHTTP method")
//	//s := h.registerHandlers()
//	//s.ServeHTTP(w, r)
//	h.Router.ServeHTTP(w, r)
//}

// GetProduct retrieves a single product from the database.
func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Get the variables from the request
	vars := mux.Vars(r)
	productId := vars["id"]
	product, err := h.ProductService.Product(r.Context(), productId)
	if err != nil {
		respondWithError(w, r, http.StatusNotFound, "productId: "+productId+" was not found.")
	} else {
		respondWithJson(w, r, http.StatusOK, product)
	}
}

// GetProducts retrieves all of the products from the database.
func (h *Handler) GetProducts(w http.ResponseWriter, r *http.Request) {
	products, err := h.ProductService.Products(r.Context())
	if err != nil {
		respondWithError(w, r, http.StatusNotFound, fmt.Sprintf("An error occured retrieving products: %v", err))
	} else {
		respondWithJson(w, r, http.StatusOK, products)
	}
}

// AddProduct adds a single product to the database.
func (h *Handler) AddProduct(w http.ResponseWriter, r *http.Request) {
	product := &catalog.Product{}
	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error decoding Json during AddProduct: %v", err))
		return
	}
	log.Debugf("The body that was posted for ProductCode: %v", product.ProductCode)
	// Add the product to the database
	err := h.ProductService.CreateProduct(r.Context(), product)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error adding product: %v", err))
	} else {
		respondWithJson(w, r, http.StatusCreated, product)
	}
}

// UpdateProduct updates a product from the database.
func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	product := &catalog.Product{}
	if err := json.NewDecoder(r.Body).Decode(product); err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error decoding Json during AddProduct: %v", err))
		return
	}
	log.Debugf("The body that was PUT for ProductCode: %v", product.ProductCode)
	// Update the product to the database
	err := h.ProductService.UpdateProduct(r.Context(), product)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, fmt.Sprintf("Error updating product: %v", err))
	} else {
		respondWithJson(w, r, http.StatusAccepted, product)
	}
}

// DeleteProduct deletes a product from the database.
func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	// Get the variables from the request
	vars := mux.Vars(r)
	productId := vars["id"]
	log.Debugf("DeleteProduct(): From the request productId=%v", productId)
	err := h.ProductService.DeleteProduct(r.Context(), productId)
	if err != nil {
		respondWithError(w, r, http.StatusNotFound, fmt.Sprintf("No product was found during delete: %v", err))
	} else {
		respondWithJson(w, r, http.StatusOK, map[string]string{"result": "success"})
	}
}

func respondWithJson(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.WithField("httpRequest", r).
			Errorf("Error marshalling Json: %v", err)
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	respondWithJson(w, r, code, map[string]string{"error": message})
}

type CustomClaims struct {
	Scope string `json:"scope"`
	jwt.StandardClaims
}

func checkScope(scope string, tokenString string) bool {
	token, _ := jwt.ParseWithClaims(tokenString, &CustomClaims{}, nil)
	claims, _ := token.Claims.(*CustomClaims)
	hasScope := false
	result := strings.Split(claims.Scope, " ")
	for i := range result {
		if result[i] == scope {
			hasScope = true
		}
	}
	return hasScope
}

func readMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authHeaderParts := strings.Split(r.Header.Get("Authorization"), " ")
	token := authHeaderParts[1]

	hasScope := checkScope("read:product", token)
	if !hasScope {
		message := "Insufficient scope. read:product needed"
		respondWithError(w, r, 401, message)
		return
	}
	// Call the next handler
	next(w, r)
}

func writeMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	authHeaderParts := strings.Split(r.Header.Get("Authorization"), " ")
	token := authHeaderParts[1]

	hasScope := checkScope("write:product", token)
	if !hasScope {
		message := "Insufficient scope. write:product needed"
		respondWithError(w, r, 401, message)
		return
	}
	// Call the next handler
	next(w, r)
}