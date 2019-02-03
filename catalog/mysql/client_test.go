package mysql

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient()
	if c.productService.client == nil {
		t.Errorf("failed to return productService client")
	}
}

