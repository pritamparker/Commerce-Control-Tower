package store

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// CartItem represents a purchasable item held in a cart or order.
type CartItem struct {
	SKU      string  `json:"sku"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// Cart groups the items pending checkout.
type Cart struct {
	Items map[string]CartItem `json:"items"`
}

// Order holds the data for a finalized purchase.
//todo: add validation for the order
type Order struct {
	ID            string     `json:"id"`
	UserID        string     `json:"userId"`
	Items         []CartItem `json:"items"`
	TotalAmount   float64    `json:"totalAmount"`
	DiscountCode  string     `json:"discountCode,omitempty"`
	DiscountValue float64    `json:"discountValue"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// DiscountCode represents a promotional code available to customers.
//todo: add validation for the discount code
type DiscountCode struct {
	Code             string    `json:"code"`
	Percentage       float64   `json:"percentage"`
	GeneratedAt      time.Time `json:"generatedAt"`
	RedeemedAt       time.Time `json:"redeemedAt,omitempty"`
	IsRedeemed       bool      `json:"isRedeemed"`
	EligibleOrderNum int       `json:"eligibleOrderNumber"`
}

// Stats aggregates store metrics for the admin dashboard.
//todo: add validation for the stats
type Stats struct {
	TotalOrders        int            `json:"totalOrders"`
	TotalItemsSold     int            `json:"totalItemsSold"`
	GrossRevenue       float64        `json:"grossRevenue"`
	TotalDiscountGiven float64        `json:"totalDiscountGiven"`
	DiscountCodes      []DiscountCode `json:"discountCodes"`
	ActiveDiscount     *DiscountCode  `json:"activeDiscount,omitempty"`
}

// MemoryStore is an in-memory implementation for the exercise.
//todo: add validation for the item
type MemoryStore struct {
	mu                sync.Mutex
	carts             map[string]*Cart
	orders            []Order
	discountHistory   []DiscountCode
	activeDiscount    *DiscountCode
	nthOrderThreshold int
	nextEligibleOrder int
	totalItemsSold    int
	grossRevenue      float64
	totalDiscount     float64
}

var (
	ErrCartEmpty           = errors.New("cart is empty")
	ErrDiscountNotActive   = errors.New("no active discount code")
	ErrDiscountAlreadyUsed = errors.New("discount code already used")
	ErrDiscountMismatch    = errors.New("discount code mismatch")
	ErrDiscountNotEligible = errors.New("not eligible to generate discount code yet")
)

// NewMemoryStore constructs a MemoryStore with sane defaults.
func NewMemoryStore(nthOrder int) *MemoryStore {
	if nthOrder <= 0 {
		nthOrder = 5
	}
	return &MemoryStore{
		carts:             make(map[string]*Cart),
		nthOrderThreshold: nthOrder,
		nextEligibleOrder: nthOrder,
	}
}

// AddItem appends or updates an item inside the user's cart.
// todo: add validation for the item
func (s *MemoryStore) AddItem(userID string, item CartItem) Cart {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart := s.getOrCreateCart(userID)
	existing, ok := cart.Items[item.SKU]
	if ok {
		existing.Quantity += item.Quantity
		existing.Price = item.Price
		existing.Name = item.Name
		cart.Items[item.SKU] = existing
	} else {
		cart.Items[item.SKU] = item
	}
	return *cart
}

// ViewCart fetches the current snapshot for the user.
func (s *MemoryStore) ViewCart(userID string) Cart {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart := s.getOrCreateCart(userID)
	return *cart
}

// Checkout finalizes the purchase and optionally redeems a discount code.
func (s *MemoryStore) Checkout(userID, discountCode string) (Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cart := s.getOrCreateCart(userID)
	if len(cart.Items) == 0 {
		return Order{}, ErrCartEmpty
	}

	var (
		items           []CartItem
		gross           float64
		discountApplied float64
		codeUsed        string
	)

	for _, item := range cart.Items {
		items = append(items, item)
		gross += float64(item.Quantity) * item.Price // total amount before discount
		s.totalItemsSold += item.Quantity // total items sold
	}

	if discountCode != "" {
		if s.activeDiscount == nil {
			return Order{}, ErrDiscountNotActive
		}
		if s.activeDiscount.IsRedeemed {
			return Order{}, ErrDiscountAlreadyUsed
		}
		if s.activeDiscount.Code != discountCode {
			return Order{}, ErrDiscountMismatch
		}
		discountApplied = gross * (s.activeDiscount.Percentage / 100)
		s.activeDiscount.IsRedeemed = true
		s.activeDiscount.RedeemedAt = time.Now()
		codeUsed = s.activeDiscount.Code
		s.discountHistory = append(s.discountHistory, *s.activeDiscount)
		s.activeDiscount = nil
		s.nextEligibleOrder += s.nthOrderThreshold
		s.totalDiscount += discountApplied // total discount given
	}

	order := Order{
		ID:            s.generateOrderID(),
		UserID:        userID,
		Items:         items,
		TotalAmount:   gross - discountApplied, // total amount after discount
		DiscountCode:  codeUsed,
		DiscountValue: discountApplied,
		CreatedAt:     time.Now(),
	}

	s.orders = append(s.orders, order)
	s.grossRevenue += gross

	// reset cart
	cart.Items = make(map[string]CartItem)

	return order, nil
}

// GenerateDiscount creates a discount code if the store is eligible.
func (s *MemoryStore) GenerateDiscount() (DiscountCode, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	totalOrders := len(s.orders)
	eligible := totalOrders >= s.nextEligibleOrder
	if !eligible || (s.activeDiscount != nil && !s.activeDiscount.IsRedeemed) {
		return DiscountCode{}, ErrDiscountNotEligible
	}

	code := DiscountCode{
		Code:             s.generateDiscountCode(),
		Percentage:       10,
		GeneratedAt:      time.Now(),
		EligibleOrderNum: s.nextEligibleOrder,
	}
	s.activeDiscount = &code
	return code, nil
}

// Stats returns aggregated data for admins.
func (s *MemoryStore) Stats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := append([]DiscountCode(nil), s.discountHistory...)
	if s.activeDiscount != nil {
		history = append(history, *s.activeDiscount)
	}

	stats := Stats{
		TotalOrders:        len(s.orders),
		TotalItemsSold:     s.totalItemsSold,
		GrossRevenue:       s.grossRevenue,
		TotalDiscountGiven: s.totalDiscount,
		DiscountCodes:      history,
	}
	if s.activeDiscount != nil {
		stats.ActiveDiscount = s.activeDiscount
	}

	return stats
}

func (s *MemoryStore) getOrCreateCart(userID string) *Cart {
	cart, ok := s.carts[userID]
	if !ok {
		cart = &Cart{Items: make(map[string]CartItem)}
		s.carts[userID] = cart
	}
	return cart
}

func (s *MemoryStore) generateOrderID() string {
	return fmt.Sprintf("ord_%d", time.Now().UnixNano())
}

func (s *MemoryStore) generateDiscountCode() string {
	const letters = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("DISC-%s", string(b))
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
