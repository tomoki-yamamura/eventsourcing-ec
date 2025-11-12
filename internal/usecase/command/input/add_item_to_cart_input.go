package input

// AddItemToCartInput represents input for adding item to cart
type AddItemToCartInput struct {
	CartID   string  `json:"cart_id"`
	ItemID   string  `json:"item_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}