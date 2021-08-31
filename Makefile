test:
	go test -v -cover -covermode=atomic ./...

unittest:
	go test -short  ./...

run:
	docker-compose up --build -d

stop:
	docker-compose down

.PHONY: unittest build docker run stop