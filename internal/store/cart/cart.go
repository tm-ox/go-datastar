package cart

type CartItem struct {
	ID        int
	CartID    string
	ProductID int
	Quantity  int
}

type CartItemDetail struct {
	ID        int
	ProductID int
	Name      string
	Price     int
	Quantity  int
}

type CartStore interface {
	GetOrCreate(cartID string) error
	AddItem(cartID string, productID int, maxStock int) error
	GetItems(cartID string) ([]CartItem, error)
	TotalQuantity(cartID string) (int, error)
	UpdateQuantity(cartID string, productID int, qty int) error
	RemoveItem(cartID string, productID int) error
	Clear(cartID string) error
	GetItemDetails(cartID string) ([]CartItemDetail, error)
}
