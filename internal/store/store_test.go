package store

import "testing"

func TestAddItemAndCheckout(t *testing.T) {
	s := NewMemoryStore(3)

	s.AddItem("user-1", CartItem{SKU: "SKU1", Name: "Widget", Price: 10, Quantity: 2})
	s.AddItem("user-1", CartItem{SKU: "SKU2", Name: "Cable", Price: 5, Quantity: 1})

	order, err := s.Checkout("user-1", "")
	if err != nil {
		t.Fatalf("checkout failed: %v", err)
	}

	if order.TotalAmount != 25 {
		t.Fatalf("expected total 25 got %v", order.TotalAmount)
	}

	cart := s.ViewCart("user-1")
	if len(cart.Items) != 0 {
		t.Fatalf("cart should be empty after checkout")
	}
}

func TestDiscountLifecycle(t *testing.T) {
	s := NewMemoryStore(1)

	s.AddItem("user-1", CartItem{SKU: "SKU1", Name: "Widget", Price: 100, Quantity: 1})
	if _, err := s.Checkout("user-1", ""); err != nil {
		t.Fatalf("initial checkout failed: %v", err)
	}

	code, err := s.GenerateDiscount()
	if err != nil {
		t.Fatalf("expected discount code, got %v", err)
	}

	s.AddItem("user-2", CartItem{SKU: "SKU2", Name: "Bag", Price: 50, Quantity: 1})
	order, err := s.Checkout("user-2", code.Code)
	if err != nil {
		t.Fatalf("checkout with discount failed: %v", err)
	}

	expected := 45.0
	if order.TotalAmount != expected {
		t.Fatalf("expected total %v got %v", expected, order.TotalAmount)
	}

	stats := s.Stats()
	if stats.TotalDiscountGiven <= 0 {
		t.Fatalf("expected discount to be tracked")
	}
}
