package mysql

import (
	"github.com/mvonbodun/go-package-test/catalog"
	log "github.com/sirupsen/logrus"
	"fmt"
	"golang.org/x/net/context"
)

// Ensure ProductService implements catalog.ProductService
var _ catalog.ProductService = &ProductService{}

// ProductService represents a service for managing Products
type ProductService struct {
	client *Client
}

// Product returns a Product by ID.
func (s *ProductService) Product(ctx context.Context, id string) (*catalog.Product, error) {
	var product catalog.Product
	// Retrieve the Product record.
	err := s.client.db.QueryRowContext(ctx, "SELECT id, productcode, shortdesc, longdesc FROM product WHERE id = ?", id).
			Scan(&product.ID, &product.ProductCode, &product.ShortDesc, &product.LongDesc)
	if err != nil {
		log.Warningf("Error retrieving product: %v, %v", id, err)
	}
	return &product, err
}

// Products returns all Products.
func (s *ProductService) Products(ctx context.Context) ([]*catalog.Product, error) {
	// Query all rows in the database
	rows, err := s.client.db.QueryContext(ctx, "SELECT id, productcode, shortdesc, longdesc FROM product")
	if err != nil {
		log.Fatalf("Error retrieving products: %v", err)
	}
	defer rows.Close()
	// Iterate over the results
	var products []*catalog.Product
	for rows.Next() {
		var product catalog.Product
		if err := rows.Scan(&product.ID, &product.ProductCode, &product.ShortDesc, &product.LongDesc); err != nil {
			log.Fatal("Error scanning over rows: %v", err)
		}
		// Add the product record to the slice
		products = append(products, &product)
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating over rows: %v", err)
	}
	return products, err
}

// CreateProduct stores a new product in the database.
func (s *ProductService) CreateProduct(ctx context.Context, product *catalog.Product) error {
	// Insert a product into the database
	stmt, err := s.client.db.PrepareContext(ctx, "INSERT product SET productcode=?, shortdesc=?, longdesc=?")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.ExecContext(ctx, product.ProductCode, product.ShortDesc, product.LongDesc)
	if err != nil {
		log.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("New product.ProductId: %v", id)
	return err
}

// DeleteProduct deletes a product in the database.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	// Delete a product from the database
	stmt, err := s.client.db.PrepareContext(ctx, "DELETE from product where id=?")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.ExecContext(ctx, id)
	if err != nil {
		log.Fatal(err)
	}
	affect, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(affect)
	return err
}