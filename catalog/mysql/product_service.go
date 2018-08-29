package mysql

import (
	"github.com/mvonbodun/go-package-test/catalog"
	log "github.com/sirupsen/logrus"
	"fmt"
)

// Ensure ProductService implements catalog.ProductService
var _ catalog.ProductService = &ProductService{}

// ProductService represents a service for managing Products
type ProductService struct {
	client *Client
}

// Product returns a Product by ID.
func (s *ProductService) Product(id string) (*catalog.Product, error) {
	var product catalog.Product
	// Retrieve the Product record.
	err := s.client.db.QueryRow("SELECT id, productcode, shortdesc, longdesc FROM product WHERE id = ?", id).
			Scan(&product.ID, &product.ProductCode, &product.ShortDesc, &product.LongDesc)
	if err != nil {
		log.Printf("Error retrieving product: %v, %v\n", id, err)
	}
	fmt.Println(product)
	return &product, err
}

// Products returns all Products.
func (s *ProductService) Products() ([]*catalog.Product, error) {
	// Query all rows in the database
	rows, err := s.client.db.Query("SELECT id, productcode, shortdesc, longdesc FROM product")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	// Iterate over the results
	var products []*catalog.Product
	for rows.Next() {
		log.Printf("About to evaluate result set: %v", rows)
		var product catalog.Product
		if err := rows.Scan(&product.ID, &product.ProductCode, &product.ShortDesc, &product.LongDesc); err != nil {
			log.Fatal(err)
		}
		fmt.Println(product)
		// Add the product record to the slice
		products = append(products, &product)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return products, err
}

// CreateProduct stores a new product in the database.
func (s *ProductService) CreateProduct(product *catalog.Product) error {
	// Insert a product into the database
	stmt, err := s.client.db.Prepare("INSERT product SET id=?, productcode=?, shortdesc=?, longdesc=?")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(product.ID, product.ProductCode, product.ShortDesc, product.LongDesc)
	if err != nil {
		log.Fatal(err)
	}
	//id, err := res.LastInsertId()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//fmt.Println(id)
	log.Printf("result: %v\n", res)
	return err
}

// DeleteProduct deletes a product in the database.
func (s *ProductService) DeleteProduct(id string) error {
	// Delete a product from the database
	stmt, err := s.client.db.Prepare("DELETE from product where id=?")
	if err != nil {
		log.Fatal(err)
	}
	res, err := stmt.Exec(id)
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