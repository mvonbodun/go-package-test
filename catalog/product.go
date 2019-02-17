package catalog

import "golang.org/x/net/context"

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
	Product(ctx context.Context, id string) (*Product, error)
	Products(ctx context.Context) ([]*Product, error)
	CreateProduct(ctx context.Context, p *Product) error
	UpdateProduct(ctx context.Context, p *Product) error
	DeleteProduct(ctx context.Context, id string) error
}
