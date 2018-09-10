package mysql

import (
	"github.com/mvonbodun/go-package-test/catalog"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"github.com/basvanbeek/ocsql"
)

type Client struct {
	// Services
	productService ProductService

	// Reference to the database
	db *sql.DB
}

func NewClient() *Client {
	c := &Client{
	}
	c.productService.client = c
	return c
}

// Open opens the connection to the MySql database
func (c *Client) Open() error {
	log.Debug("Before opening the database")
	// Setup the OpenCensus database tracing
	ocDriverName, err := ocsql.Register("mysql", ocsql.WithAllTraceOptions())
	if err != nil {
		log.Fatalf("Failed to register the ocsql driver: %v", err)
	}
	db, err := sql.Open(ocDriverName, "root:passw0rd@/catalog?charset=utf8")
	if err != nil {
		log.Fatalf("Failed to open the catalog Database: %v",err)
	}
	c.db = db
	// Ping the database
	log.Debug("Before pinging mysql catalog database.")
	err = db.Ping()
	if err != nil {
		log.Fatalf("Could not ping the catalog database: %v\n", err)
	}
	// Prepare the SQL statements
	err = c.productService.prepareSqlStmts()
	if err != nil {
		log.Errorf("mysql client: Failed to prepare sql statements: %v", err)
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
