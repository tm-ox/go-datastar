package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
	"github.com/tm-ox/go-datastar/internal/content"
	"github.com/tm-ox/go-datastar/views/modules"
	views "github.com/tm-ox/go-datastar/views/pages"

	"github.com/a-h/templ"
)

type route struct {
	Path    string
	Label   string
	Handler http.HandlerFunc
}

func main() {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))

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

	var navItems []modules.NavItem

	routes := []route{
		{Path: "/", Label: "Home", Handler: func(w http.ResponseWriter, r *http.Request) {
			templ.Handler(views.Index(navItems, "/", site.Home)).ServeHTTP(w, r)
		}},
		{Path: "/about", Label: "About", Handler: func(w http.ResponseWriter, r *http.Request) {
			templ.Handler(views.About(navItems, "/about", site.About)).ServeHTTP(w, r)
		}},
		{Path: "/work", Label: "Work", Handler: func(w http.ResponseWriter, r *http.Request) {
			types := content.UniqueTypes(workEntries)
			clients := content.UniqueClients(workEntries)
			years := content.UniqueYears(workEntries)
			tools := content.UniqueTools(workEntries)
			templ.Handler(views.WorkIndex(navItems, "/work", workEntries, types, clients, years, tools)).ServeHTTP(w, r)
		}},
		{Path: "/work/{slug}", Handler: func(w http.ResponseWriter, r *http.Request) {
			slug := r.PathValue("slug")
			entry, ok := workMap[slug]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			templ.Handler(views.WorkDetail(navItems, r.URL.Path, entry)).ServeHTTP(w, r)
		}},
		{Path: "/work/filter", Handler: func(w http.ResponseWriter, r *http.Request) {
			var sig struct {
				Type   string `json:"type"`
				Client string `json:"client"`
				Year   string `json:"year"`
				Tool   string `json:"tool"`
			}
			if err := datastar.ReadSignals(r, &sig); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			filtered := content.FilterWork(workEntries, sig.Type, sig.Client, sig.Year, sig.Tool)
			sse := datastar.NewSSE(w, r)
			sse.PatchElementTempl(views.WorkRows(filtered))

		}},
	}

	for _, r := range routes {
		mux.HandleFunc(r.Path, r.Handler)
	}

	for _, r := range routes {
		if r.Label != "" {
			navItems = append(navItems, modules.NavItem{Label: r.Label, URL: r.Path})
		}
	}

	fmt.Println("Listening on :8081")
	http.ListenAndServe(":8081", mux)
}
