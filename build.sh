gofmt -w .
go build -o ./build/ ./cmd/pages
./build/pages test2.db