package input

type AddItemToCartInput struct {
	CartID   string  `json:"cart_id"`
	UserID   string  `json:"user_id"`
	ItemID   string  `json:"item_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}
