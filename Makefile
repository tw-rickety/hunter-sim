
run:
	go run cmd/main.go

test:
	go test -v ./...

build: 
	go build -o cmd/main cmd/main.go
