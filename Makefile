.PHONY: build test clean install lint

build:
	go build -o err

test:
	go test -v -race -cover

clean:
	rm -f err err.exe
	rm -f coverage.txt

install:
	go install

lint:
	golangci-lint run

bench:
	go test -bench=. -benchmem

coverage:
	go test -coverprofile=coverage.txt -covermode=atomic
	go tool cover -html=coverage.txt
