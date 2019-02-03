package mock

import (
	"github.com/mvonbodun/go-package-test/catalog"
	"golang.org/x/net/context"
)

type ProductService struct {
	ProductFn 		func(ctx context.Context, id string) (*catalog.Product, error)
	ProductInvoked bool

	ProductsFn 		func(ctx context.Context) ([]*catalog.Product, error)
	ProductsInvoked bool

	CreateProductFn func(ctx context.Context, product *catalog.Product) error
	CreateProductInvoked bool

	DeleteProductFn func(ctx context.Context, id string) error
	DeleteProductInvoked bool
}

func (s *ProductService) Product(ctx context.Context, id string) (*catalog.Product, error) {
	s.ProductInvoked = true
	return s.ProductFn(context.Background(), id)
}

func (s *ProductService) Products(ctx context.Context) ([]*catalog.Product, error) {
	s.ProductsInvoked = true
	return s.ProductsFn(context.Background())
}

func (s *ProductService) CreateProduct(ctx context.Context, product *catalog.Product) error {
	s.CreateProductInvoked = true
	return s.CreateProductFn(context.Background(), product)
}

func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	s.DeleteProductInvoked = true
	return s.DeleteProductFn(context.Background(), id)
}