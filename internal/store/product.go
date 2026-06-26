package store

type Product struct {
	Image             string
	Name              string
	Description       string
	Price             int
	Category          string
	Slug              string
	CompareAtPrice    int
	AvailableForSale  bool
	QuantityAvailable int
	Vendor            string
}

type ProductQuery struct {
	ProductType string
	Vendor      string
	InStock     bool
	Search      string
	After       string
	Before      string
	First       int
	Last        int
	Page        int
}
