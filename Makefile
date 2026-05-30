.PHONY: dev build css templ clean

dev:
	$(MAKE) -j4 templ css run sync

templ:
	templ generate --watch

sync:
	bun run sync

css:
	bun run watch:css

run:
	air

build:
	templ generate
	bun run build:css
	go build -o bin/server ./cmd/main.go

clean:
	rm -rf bin/ static/app.css views/*/*_templ.go
