package shopify

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Khan/genqlient/graphql"
	"github.com/tm-ox/go-datastar/internal/store"
)

const collectionHandle = "shop-all-products"

type ProductStore struct {
	client graphql.Client
}

func NewProductStore(domain, token string) *ProductStore {
	endpoint := fmt.Sprintf("https://%s/api/2025-01/graphql.json", domain)
	httpClient := &http.Client{
		Transport: &tokenTransport{token: token},
	}
	return &ProductStore{client: graphql.NewClient(endpoint, httpClient)}
}

type tokenTransport struct {
	token string
}

func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("X-Shopify-Storefront-Access-Token", t.token)
	return http.DefaultTransport.RoundTrip(req)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func parseCents(amount string) int {
	f, err := strconv.ParseFloat(strings.TrimSpace(amount), 64)
	if err != nil {
		return 0
	}
	return int(f * 100)
}

func (s *ProductStore) List(q store.ProductQuery) ([]store.Product, bool, bool, string, string, int, error) {
	first := q.First
	if first == 0 {
		first = 20
	}
	resp, err := ListProducts(context.Background(), s.client, collectionHandle, &first, nil, nil, nil, nil)
	if err != nil {
		log.Printf("shopify List error: %v", err)
		return nil, false, false, "", "", 0, err
	}
	return mapProducts(resp.Collection.Products.Nodes),
		resp.Collection.Products.PageInfo.HasNextPage,
		resp.Collection.Products.PageInfo.HasPreviousPage,
		resp.Collection.Products.PageInfo.EndCursor,
		resp.Collection.Products.PageInfo.StartCursor,
		extractTotal(resp.Collection.Products.Filters, false),
		nil
}

func (s *ProductStore) Filter(q store.ProductQuery) ([]store.Product, bool, bool, string, string, int, error) {
	if q.ProductType != "" || q.Vendor != "" || q.InStock || q.Search != "" {
		return s.clientFilter(q)
	}

	var (
		resp *ListProductsResponse
		err  error
	)
	if q.Last > 0 {
		resp, err = ListProducts(context.Background(), s.client, collectionHandle, nil, nil, &q.Last, strPtr(q.Before), nil)
	} else {
		first := q.First
		if first == 0 {
			first = 20
		}
		resp, err = ListProducts(context.Background(), s.client, collectionHandle, &first, strPtr(q.After), nil, nil, nil)
	}
	if err != nil {
		return nil, false, false, "", "", 0, err
	}
	return mapProducts(resp.Collection.Products.Nodes),
		resp.Collection.Products.PageInfo.HasNextPage,
		resp.Collection.Products.PageInfo.HasPreviousPage,
		resp.Collection.Products.PageInfo.EndCursor,
		resp.Collection.Products.PageInfo.StartCursor,
		extractTotal(resp.Collection.Products.Filters, false),
		nil
}

func (s *ProductStore) clientFilter(q store.ProductQuery) ([]store.Product, bool, bool, string, string, int, error) {
	first := 250
	resp, err := ListProducts(context.Background(), s.client, collectionHandle, &first, nil, nil, nil, nil)
	if err != nil {
		return nil, false, false, "", "", 0, err
	}
	term := strings.ToLower(q.Search)
	all := mapProducts(resp.Collection.Products.Nodes)
	matched := make([]store.Product, 0, len(all))
	for _, p := range all {
		if q.InStock && !p.AvailableForSale {
			continue
		}
		if q.ProductType != "" && p.Category != q.ProductType {
			continue
		}
		if q.Vendor != "" && p.Vendor != q.Vendor {
			continue
		}
		if term != "" && !strings.Contains(strings.ToLower(p.Name), term) &&
			!strings.Contains(strings.ToLower(p.Description), term) &&
			!strings.Contains(strings.ToLower(p.Vendor), term) {
			continue
		}
		matched = append(matched, p)
	}
	total := len(matched)
	limit := q.First
	if limit <= 0 {
		limit = 9
	}
	page := q.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit
	if offset >= total {
		offset = 0
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return matched[offset:end], end < total, page > 1, "", "", total, nil
}

func extractTotal(filters []ListProductsCollectionProductsProductConnectionFiltersFilter, inStock bool) int {
	for _, f := range filters {
		if f.Id != "filter.v.availability" {
			continue
		}
		if !inStock {
			total := 0
			for _, v := range f.Values {
				total += v.Count
			}
			return total
		}
		for _, v := range f.Values {
			if strings.Contains(v.Input, `"available":true`) {
				return v.Count
			}
		}
	}
	return 0
}

func (s *ProductStore) GetByHandle(handle string) (*store.Product, error) {
	resp, err := GetProduct(context.Background(), s.client, handle)
	if err != nil {
		return nil, err
	}
	if resp.Product.Handle == "" {
		return nil, nil
	}
	p := resp.Product
	var image string
	if len(p.Images.Nodes) > 0 {
		image = p.Images.Nodes[0].Url
	}
	return &store.Product{
		Name:              p.Title,
		Description:       p.Description,
		Slug:              p.Handle,
		Category:          p.ProductType,
		Vendor:            p.Vendor,
		Price:             parseCents(p.PriceRange.MinVariantPrice.Amount),
		CompareAtPrice:    parseCents(p.CompareAtPriceRange.MinVariantPrice.Amount),
		AvailableForSale:  p.AvailableForSale,
		QuantityAvailable: p.TotalInventory,
		Image:             image,
	}, nil
}

func (s *ProductStore) FilterMeta() (productTypes []string, vendors []string, err error) {
	first := 250
	resp, err := ListProducts(context.Background(), s.client, collectionHandle, &first, nil, nil, nil, nil)
	if err != nil {
		return nil, nil, err
	}
	seenTypes := map[string]bool{}
	seenVendors := map[string]bool{}
	for _, n := range resp.Collection.Products.Nodes {
		if n.ProductType != "" && !seenTypes[n.ProductType] {
			seenTypes[n.ProductType] = true
			productTypes = append(productTypes, n.ProductType)
		}
		if n.Vendor != "" && !seenVendors[n.Vendor] {
			seenVendors[n.Vendor] = true
			vendors = append(vendors, n.Vendor)
		}
	}
	return productTypes, vendors, nil
}

func mapProducts(nodes []ListProductsCollectionProductsProductConnectionNodesProduct) []store.Product {
	products := make([]store.Product, 0, len(nodes))
	for _, n := range nodes {
		var image string
		if len(n.Images.Nodes) > 0 {
			image = n.Images.Nodes[0].Url
		}
		products = append(products, store.Product{
			Name:              n.Title,
			Description:       n.Description,
			Slug:              n.Handle,
			Category:          n.ProductType,
			Vendor:            n.Vendor,
			Price:             parseCents(n.PriceRange.MinVariantPrice.Amount),
			CompareAtPrice:    parseCents(n.CompareAtPriceRange.MinVariantPrice.Amount),
			AvailableForSale:  n.AvailableForSale,
			QuantityAvailable: n.TotalInventory,
			Image:             image,
		})
	}
	return products
}
