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
	"github.com/tm-ox/go-datastar/internal/store/product"
	"github.com/tm-ox/go-datastar/views/modules"
)

func main() {
	site, err := content.Load()
	if err != nil {
		log.Fatalf("failed to load content: %v", err)
	}

	workEntries, err := content.LoadWork()
	if err != nil {
		log.Fatalf("failed to load work: %v", err)
	}

	workMap := make(map[string]content.WorkEntry, len(workEntries))
	for _, e := range workEntries {
		workMap[e.Slug] = e
	}

	nav := []modules.NavItem{
		{Label: "Home", URL: "/"},
		{Label: "About", URL: "/about"},
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

	productStore := product.NewSQLiteProductStore(database)

	site_h := handler.NewSiteHandler(nav, site)
	work_h := handler.NewWorkHandler(nav, workEntries, workMap)
	shop_h := handler.NewShopHandler(nav, productStore)
	settings_h := handler.NewSettingsHandler(nav, settingsSections)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", site_h.Index)
	mux.HandleFunc("/about", site_h.About)
	mux.HandleFunc("/work", work_h.Index)
	mux.HandleFunc("/work/{slug}", work_h.Detail)
	mux.HandleFunc("/work/filter", work_h.Filter)
	mux.HandleFunc("/shop", shop_h.Index)
	mux.HandleFunc("/shop/{slug}", shop_h.Detail)
	mux.HandleFunc("/shop/filter", shop_h.Filter)
	mux.HandleFunc("/settings", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/settings/work", http.StatusFound)
	})
	mux.HandleFunc("/settings/work", settings_h.Work)
	mux.HandleFunc("/settings/shop", settings_h.Shop)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8081", Handler: middleware.Logging(mux)}
	fmt.Println("Listening on :8081")
	go srv.ListenAndServe()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
