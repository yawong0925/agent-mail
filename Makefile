.PHONY: build run clean

build:
	go build -o bin/agent-mail main.go

run: build
	./bin/agent-mail

clean:
	rm -rf bin/ data/