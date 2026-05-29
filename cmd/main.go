package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/internal/handler"
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
	}

	site_h := handler.NewSiteHandler(nav, site)
	work_h := handler.NewWorkHandler(nav, workEntries, workMap)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", site_h.Index)
	mux.HandleFunc("/about", site_h.About)
	mux.HandleFunc("/work", work_h.Index)
	mux.HandleFunc("/work/{slug}", work_h.Detail)
	mux.HandleFunc("/work/filter", work_h.Filter)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{Addr: ":8081", Handler: mux}
	fmt.Println("Listening on :8081")
	go srv.ListenAndServe()
	<-ctx.Done()
	srv.Shutdown(context.Background())
}
