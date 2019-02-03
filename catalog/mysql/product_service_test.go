package mysql

import (
	"context"
	"fmt"
	"github.com/mvonbodun/go-package-test/catalog"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"log"
	"testing"
)



func TestProductService_Product(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	columns := []string{"id", "productcode", "shortdesc", "longdesc"}
	mock.ExpectPrepare("SELECT id, productcode, shortdesc, longdesc FROM product WHERE id = \\?").
		ExpectQuery().WithArgs("5").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("5", "1234", "shortdesc for 1234", "longdesc for 1234"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(getstmt)
	_, err = client.productService.Product(context.Background(), "5")
	if err != nil {
		t.Errorf("expected no error, but got %s instead", err)
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_ProductNotFound(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("SELECT id, productcode, shortdesc, longdesc FROM product WHERE id = \\?").
		ExpectQuery().WithArgs("5").
		WillReturnError(fmt.Errorf("no results"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(getstmt)
	product, err := client.productService.Product(context.Background(), "5")
	if err == nil {
		t.Errorf("expected error, but got none")
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	log.Printf("productcode: %v", product.ProductCode)
}

func TestProductService_Products(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()


	columns := []string{"id", "productcode", "shortdesc", "longdesc"}
	mock.ExpectPrepare("SELECT id, productcode, shortdesc, longdesc FROM product").
		ExpectQuery().
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("5", "1234", "shortdesc for 1234", "longdesc for 1234").
			AddRow("6", "5678", "shortdesc for 5678", "longdesc for 5678"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(liststmt)
	_, err = client.productService.Products(context.Background())
	if err != nil {
		t.Errorf("expected no error, but got %s instead", err)
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_ProductsNoneReturned(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()


	//columns := []string{"id", "productcode", "shortdesc", "longdesc"}
	mock.ExpectPrepare("SELECT id, productcode, shortdesc, longdesc FROM product").
		ExpectQuery().
		WillReturnError(fmt.Errorf("no results"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(liststmt)
	_, err = client.productService.Products(context.Background())
	if err == nil {
		t.Errorf("expected error, but got none")
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_ProductErrorScanningRows(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	columns := []string{"id", "productcode", "shortdesc", "longdesc"}
	mock.ExpectPrepare("SELECT id, productcode, shortdesc, longdesc FROM product").
		ExpectQuery().
		WillReturnRows(sqlmock.NewRows(columns).RowError(1, fmt.Errorf("error reading row")).
			AddRow("5", "1234", "shortdesc for 1234", "longdesc for 1234").
			AddRow("6", "5678", "shortdesc for 5678", "longdesc for 5678"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(liststmt)
	_, err = client.productService.Products(context.Background())
	if err == nil {
		t.Errorf("expected error, but got none")
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("INSERT product SET productcode=\\?, shortdesc=\\?, longdesc=\\?").
		ExpectExec().
		WithArgs("1234", "shortdesc for 1234", "longdesc for 1234").
		WillReturnResult(sqlmock.NewResult(1, 1))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(insertstmt)
	product := &catalog.Product{
		ProductCode: "1234",
		ShortDesc: "shortdesc for 1234",
		LongDesc: "longdesc for 1234",
	}
	// execute the method
	if err := client.productService.CreateProduct(context.Background(), product); err != nil {
		t.Errorf("error was not expected while updating stats: %s", err)
	}
	// makes sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_CreateProductFailedInsert(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("INSERT product SET productcode=\\?, shortdesc=\\?, longdesc=\\?").
		ExpectExec().
		WillReturnError(fmt.Errorf("error inserting row"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(insertstmt)
	product := &catalog.Product{
		ProductCode: "1234",
		ShortDesc: "shortdesc for 1234",
		LongDesc: "longdesc for 1234",
	}

	if err := client.productService.CreateProduct(context.Background(), product); err == nil {
		t.Errorf("error expected but got none")
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("DELETE from product where id=\\?").
		ExpectExec().
		WillReturnResult(sqlmock.NewResult(0, 1))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(deletestmt)

	if err := client.productService.DeleteProduct(context.Background(), "1"); err != nil {
		t.Errorf("expected no error but got: %v instead", err)
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestProductService_DeleteProductFailedDelete (t *testing.T) {
	// Create DB Mock
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectPrepare("DELETE from product where id=\\?").
		ExpectExec().
		WillReturnError(fmt.Errorf("failed deleting record"))

	client := NewClient()
	client.db = db
	client.productService.prepareSqlStmt(deletestmt)

	if err := client.productService.DeleteProduct(context.Background(), "1"); err == nil {
		t.Errorf("expected error but got none")
	}
	// make sure expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

}