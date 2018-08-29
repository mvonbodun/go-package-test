package catalog

// ProductID represents a product identifier.
// type ProductID string

// Product represents a product for sale.
type Product struct {
	ID          string `json:"productId"`
	ProductCode string `json:"productCode"`
	ShortDesc   string `json:"shortDesc"`
	LongDesc    string `json:"longDesc"`
}

// Client creates a connection to the service.
type Client interface {
	ProductService() ProductService
}

// ProductService represents a service for managing products.
type ProductService interface {
	Product(id string) (*Product, error)
	Products() ([]*Product, error)
	CreateProduct(p *Product) error
	DeleteProduct(id string) error
}
