.PHONY: dev build css templ clean

dev:
	$(MAKE) -j4 templ css run sync

templ:
	templ generate --watch --proxy="http://localhost:8081" --open-browser=false

sync:
	bun run sync

css:
	bun run watch:css

run:
	go run ./cmd/main.go

build:
	templ generate
	bun run build:css
	go build -o bin/server ./cmd/main.go

clean:
	rm -rf bin/ static/app.css views/*/*_templ.go
