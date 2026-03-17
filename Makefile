build:
	go build -o doug-plan .

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f doug-plan

.PHONY: build test lint clean
