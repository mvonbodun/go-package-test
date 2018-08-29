package mysql

import (
	"github.com/mvonbodun/go-package-test/catalog"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	// Services
	productService ProductService

	// Reference to the database
	db *sql.DB
}

func NewClient() *Client {
	c := &Client{}
	c.productService.client = c
	return c
}

// Open opens the connection to the MySql database
func (c *Client) Open() error {
	log.Info("Before opening the database")
	db, err := sql.Open("mysql", "root:passw0rd@/catalog?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}
	c.db = db
	// Ping the database
	log.Info("Before pinging mysql.")
	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping db: %v\n", err)
	}
	return err
}

// Close closes the underlying MySql database
func (c *Client) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}


// ProductService returns the product service associated with the client
func (c *Client) ProductService() catalog.ProductService {
	return &c.productService
}
