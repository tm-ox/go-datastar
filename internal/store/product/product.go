package product

type Product struct {
	ID          int
	Name        string
	Description string
	Price       int
	Category    string
	Slug        string
	CreatedAt   string
}

type ProductStore interface {
	List() ([]Product, error)
	GetBySlug(slug string) (*Product, error)
}
