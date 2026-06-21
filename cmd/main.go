package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/internal/db"
	"github.com/tm-ox/go-datastar/internal/handler"
	"github.com/tm-ox/go-datastar/internal/middleware"
	"github.com/tm-ox/go-datastar/internal/store"
	"github.com/tm-ox/go-datastar/internal/stream"
	"github.com/tm-ox/go-datastar/views/modules"
)

func main() {
	site, err := content.Load()
	if err != nil {
		log.Fatalf("failed to load content: %v", err)
	}

	nav := []modules.NavItem{
		{Label: "Home", URL: "/"},
		// {Label: "About", URL: "/about"},
		{Label: "Context", URL: "/context"},
		{Label: "Dash", URL: "/dashboard"},
		{Label: "Work", URL: "/work"},
		{Label: "Shop", URL: "/shop"},
	}

	settingsSections := []modules.NavItem{
		{Label: "Work", URL: "/settings/work"},
		{Label: "Shop", URL: "/settings/shop"},
	}

	database, err := db.Open("./data.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	productStore := store.NewProductStore(database)
	workStore := store.NewWorkStore(database)
	cartStore := store.NewCartStore(database)
	orderStore := store.NewOrderStore(database)
	cart_h := handler.NewCartHandler(nav, cartStore, productStore, orderStore)

	hub := stream.NewHub()
	agg := stream.NewAggregator()
	src := stream.NewSource(hub, agg, "go-datastar-dashboard/1.0 (tim@tmox.net)")

	site_h := handler.NewSiteHandler(nav, site)

	dashboard_h := handler.NewDashboardHandler(nav, hub, agg)
	work_h := handler.NewWorkHandler(nav, workStore)
	shop_h := handler.NewShopHandler(nav, productStore)
	settings_h := handler.NewSettingsHandler(nav, settingsSections, productStore, workStore)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", site_h.Index)
	mux.HandleFunc("/about", site_h.About)
	mux.HandleFunc("/context", site_h.Context)
	mux.HandleFunc("/dashboard", dashboard_h.Index)
	mux.HandleFunc("GET /dashboard/stream", dashboard_h.Stream)
	mux.HandleFunc("/work", work_h.Index)
	mux.HandleFunc("/work/{slug}", work_h.Detail)
	mux.HandleFunc("/work/filter", work_h.Filter)
	mux.HandleFunc("/shop", shop_h.Index)
	mux.HandleFunc("/shop/{slug}", shop_h.Detail)
	mux.HandleFunc("/shop/filter", shop_h.Filter)
	mux.HandleFunc("POST /cart/add", cart_h.Add)
	mux.HandleFunc("/cart/total", cart_h.Total)
	mux.HandleFunc("/cart/drawer", cart_h.Drawer)
	mux.HandleFunc("POST /cart/remove", cart_h.Remove)
	mux.HandleFunc("GET /cart", cart_h.Checkout)
	mux.HandleFunc("POST /cart/qty", cart_h.UpdateQty)
	mux.HandleFunc("POST /cart/checkout", cart_h.PlaceOrder)
	mux.HandleFunc("GET /cart/success", cart_h.Success)
	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/settings/work", http.StatusFound)
	})
	mux.HandleFunc("/settings/work", settings_h.Work)
	mux.HandleFunc("/settings/work/filter", settings_h.WorkFilter)
	mux.HandleFunc("/settings/work/form", settings_h.WorkForm)
	mux.HandleFunc("POST /settings/work/create", settings_h.WorkCreate)
	mux.HandleFunc("POST /settings/work/update", settings_h.WorkUpdate)
	mux.HandleFunc("POST /settings/work/delete", settings_h.WorkDelete)
	mux.HandleFunc("/settings/shop", settings_h.Shop)
	mux.HandleFunc("/settings/shop/filter", settings_h.ShopFilter)
	mux.HandleFunc("/settings/shop/stock", settings_h.ShopStock)
	mux.HandleFunc("/settings/shop/products/form", settings_h.ShopProductForm)
	mux.HandleFunc("POST /settings/shop/products/create", settings_h.ShopProductCreate)
	mux.HandleFunc("POST /settings/shop/products/update", settings_h.ShopProductUpdate)
	mux.HandleFunc("POST /settings/shop/products/delete", settings_h.ShopProductDelete)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go src.Run(ctx)

	srv := &http.Server{Addr: ":8081", Handler: middleware.CartTotal(cartStore, middleware.Logging(mux))}
	fmt.Println("Listening on :8081")
	go srv.ListenAndServe()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
