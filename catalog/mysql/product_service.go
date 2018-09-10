package mysql

import (
	"github.com/mvonbodun/go-package-test/catalog"
	log "github.com/sirupsen/logrus"
	"fmt"
	"golang.org/x/net/context"
	"database/sql"
)

// Ensure ProductService implements catalog.ProductService
var _ catalog.ProductService = &ProductService{}

// ProductService represents a service for managing Products
type ProductService struct {
	client *Client
	get 	*sql.Stmt
	list 	*sql.Stmt
	insert 	*sql.Stmt
	delete 	*sql.Stmt
}

// prepareSqlStmts prepares the SQL statements ahead of time resulting in faster performance.
func (s *ProductService) prepareSqlStmts() error {
	var err error
	if s.get, err = s.client.db.Prepare(getstmt); err != nil {
		return fmt.Errorf("mysql: prepare get: %v", err)
	}
	if s.list, err = s.client.db.Prepare(liststmt); err != nil {
		return fmt.Errorf("mysql: prepare list: %v", err)
	}
	if s.insert, err = s.client.db.Prepare(insertstmt); err != nil {
		return fmt.Errorf("mysql: prepare insert: %v", err)
	}
	if s.delete, err = s.client.db.Prepare(deletestmt); err != nil {
		return fmt.Errorf("mysql: prepare delete: %v", err)
	}
	return nil
}

const getstmt = "SELECT id, productcode, shortdesc, longdesc FROM product WHERE id = ?"

// Product returns a Product by ID.
func (s *ProductService) Product(ctx context.Context, id string) (*catalog.Product, error) {
	var product catalog.Product
	// Retrieve the Product record.
	err := s.get.QueryRowContext(ctx, id).
		Scan(&product.ID, &product.ProductCode, &product.ShortDesc, &product.LongDesc)

	if err != nil {
		log.WithField("ctx", ctx).Warningf("Error retrieving product: %v, %v", id, err)
	}
	return &product, err
}

const liststmt = "SELECT id, productcode, shortdesc, longdesc FROM product"

// Products returns all Products.
func (s *ProductService) Products(ctx context.Context) ([]*catalog.Product, error) {
	// Query all rows in the database
	rows, err := s.list.QueryContext(ctx)
	//rows, err := s.client.db.QueryContext(ctx, "SELECT id, productcode, shortdesc, longdesc FROM product")
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

const insertstmt = "INSERT product SET productcode=?, shortdesc=?, longdesc=?"

// CreateProduct stores a new product in the database.
func (s *ProductService) CreateProduct(ctx context.Context, product *catalog.Product) error {
	// Insert a product into the database
	//stmt, err := s.client.db.PrepareContext(ctx, "INSERT product SET productcode=?, shortdesc=?, longdesc=?")
	//if err != nil {
	//	log.Fatal(err)
	//}
	res, err := s.insert.ExecContext(ctx, product.ProductCode, product.ShortDesc, product.LongDesc)
	//res, err := stmt.ExecContext(ctx, product.ProductCode, product.ShortDesc, product.LongDesc)
	if err != nil {
		log.Fatal(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}
	log.WithField("context", ctx).
		Infof("New product.ProductId: %v", id)
	return err
}

const deletestmt = "DELETE from product where id=?"

// DeleteProduct deletes a product in the database.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	// Delete a product from the database
	//stmt, err := s.client.db.PrepareContext(ctx, "DELETE from product where id=?")
	//if err != nil {
	//	log.Fatal(err)
	//}
	res, err := s.delete.ExecContext(ctx, id)
	//res, err := stmt.ExecContext(ctx, id)
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