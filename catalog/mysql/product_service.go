package mysql

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/mvonbodun/go-package-test/catalog"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"strconv"
)

// Ensure ProductService implements catalog.ProductService
var _ catalog.ProductService = &ProductService{}

// Table creation query
var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS catalog DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE catalog;`,
	`CREATE TABLE IF NOT EXISTS product (
		id INT UNSIGNED NOT NULL AUTO_INCREMENT,
		productcode VARCHAR(255) NULL,
		shortdesc VARCHAR(255) NULL,
		longdesc text NULL,
		PRIMARY KEY (id)
	)`,
}

// MySQLConfig holds the connection info for the database.
type MySQLConfig struct {
	Username, Password string
	Host string //host:port, i.e. localhost:3306
}

// ProductService represents a service for managing Products
type ProductService struct {
	client *Client
	get    *sql.Stmt
	list   *sql.Stmt
	insert *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
}

// Define custom types for statements to help with sqlmock tests
type (
	SqlStatement    string
	GetStatement    SqlStatement
	ListStatement   SqlStatement
	InsertStatement SqlStatement
	UpdateStatement SqlStatement
	DeleteStatement SqlStatement
)

// prepareSqlStmts prepares the SQL statements ahead of time resulting in faster performance.
func (s *ProductService) prepareSqlStmts() error {
	// Prepare all the SQL statements
	if err := s.prepareSqlStmt(getstmt, liststmt, insertstmt, updatestmt, deletestmt); err != nil {
		return err
	}
	return nil
}

// prepareSqlStmt is used to only prepare SQL statements due to an issue with sqlmock
// not supporting more than one prepared statement at a time
func (s *ProductService) prepareSqlStmt(stmts ...interface{}) error {
	var err error
	for _, v := range stmts {
		switch v.(type) {
		case GetStatement:
			if s.get, err = s.client.db.Prepare(string(getstmt)); err != nil {
				return fmt.Errorf("mysql: prepare get: %v", err)
			}
		case ListStatement:
			if s.list, err = s.client.db.Prepare(string(liststmt)); err != nil {
				return fmt.Errorf("mysql: prepare list: %v", err)
			}
		case InsertStatement:
			if s.insert, err = s.client.db.Prepare(string(insertstmt)); err != nil {
				return fmt.Errorf("mysql: prepare insert: %v", err)
			}
		case UpdateStatement:
			if s.update, err = s.client.db.Prepare(string(updatestmt)); err != nil {
				return fmt.Errorf("mysql: prepare update: %v", err)
			}
		case DeleteStatement:
			if s.delete, err = s.client.db.Prepare(string(deletestmt)); err != nil {
				return fmt.Errorf("mysql: prepare delete: %v", err)
			}
		}
	}
	return nil
}

var getstmt GetStatement = "SELECT id, productcode, shortdesc, longdesc FROM product WHERE id = ?"

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

var liststmt ListStatement = "SELECT id, productcode, shortdesc, longdesc FROM product"

// Products returns all Products.
func (s *ProductService) Products(ctx context.Context) ([]*catalog.Product, error) {
	// Query all rows in the database
	rows, err := s.list.QueryContext(ctx)
	if err != nil {
		log.Errorf("Error retrieving products: %v", err)
		return nil, err
	}
	defer rows.Close()
	// Iterate over the results
	var products []*catalog.Product
	for rows.Next() {
		var product catalog.Product
		if err := rows.Scan(&product.ID, &product.ProductCode, &product.ShortDesc, &product.LongDesc); err != nil {
			log.Errorf("Error scanning over rows: %v", err)
			return nil, err
		}
		// Add the product record to the slice
		products = append(products, &product)
	}
	if err := rows.Err(); err != nil {
		log.Errorf("Error iterating over rows: %v", err)
		return nil, err
	}
	return products, err
}

var insertstmt InsertStatement = "INSERT product SET productcode=?, shortdesc=?, longdesc=?"

// CreateProduct stores a new product in the database.
func (s *ProductService) CreateProduct(ctx context.Context, product *catalog.Product) error {
	res, err := s.insert.ExecContext(ctx, product.ProductCode, product.ShortDesc, product.LongDesc)
	if err != nil {
		log.Error(err)
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Error(err)
		return err
	}
	product.ID = strconv.Itoa(int(id))
	log.WithField("productId", product.ID).
		Debugf("New product.ProductId: %d", id)
	return err
}

var updatestmt UpdateStatement = "UPDATE product SET productcode=?, shortdesc=?, longdesc=? WHERE id=?"

// UpdateProduct updates an existing product in the database.
func (s *ProductService) UpdateProduct(ctx context.Context, product *catalog.Product) error {
	log.Infof("product: %v", product)
	if len(product.ID) == 0 {
		return errors.New("mysql: product with unassigned ID passed in to UpdateProduct")
	}
	res, err := s.update.ExecContext(ctx, product.ProductCode, product.ShortDesc, product.LongDesc, product.ID)
	if err != nil {
		log.Error(err)
		return err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Number of product rows updated: %d", affect)
	return nil
}

var deletestmt DeleteStatement = "DELETE from product where id=?"

// DeleteProduct deletes a product in the database.
func (s *ProductService) DeleteProduct(ctx context.Context, id string) error {
	res, err := s.delete.ExecContext(ctx, id)
	if err != nil {
		log.Error(err)
		return err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debugf("Number of product rows deleted: %d", affect)
	return nil
}

// ensureTableExists checks the table exists. If not, it creates it.
func (config MySQLConfig) ensureTableExists() error {
	mc := mysql.NewConfig()
	mc.User = config.Username
	mc.Passwd = config.Password
	mc.Net = "tcp"
	mc.Addr = config.Host
	mc.Params = map[string]string{"charset": "utf8"}
	log.Infof("ensureTableExists formatDSN: %v", mc.FormatDSN())
	conn, err := sql.Open("mysql", mc.FormatDSN())
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database. " +
			"could be bad address, or this address is not whitelisted for access.")
	}

	if _, err := conn.Exec("USE catalog"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}
	}

	if _, err := conn.Exec("DESCRIBE product"); err != nil {
		// MySQL error 1146 is "table does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// createTable creates the table, and if necessary, the database.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}
