build:
	templ generate
	mkdir build
	go build -o build

clean:
	rm -r build

lint:
	gofumpt -l -w .
	golangci-lint run -c .golangci-lint.yaml

	go mod tidy
	go clean

install-linters:
	go install mvdan.cc/gofumpt@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

run:
	templ generate
	go run .