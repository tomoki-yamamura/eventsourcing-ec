package input

// AddItemToCartInput represents input for adding item to cart
type AddItemToCartInput struct {
	CartID   string  `json:"cart_id,omitempty"` // Optional - will be generated if empty
	UserID   string  `json:"user_id"`           // Required - identifies the user
	ItemID   string  `json:"item_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}