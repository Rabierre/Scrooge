build:
	go build -o build/scrooge
clean:
	rm -rf build/
test:
	go test ./...
run:
	go run main.go
install:
	go get -v github.com/rabierre/scrooge

.PHONY: build clean test run install
