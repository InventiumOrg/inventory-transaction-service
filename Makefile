SERVICE_NAME ?= inventory-transaction-service
PORT         ?= 14330

.PHONY: build run tidy test vet lint clean docker docker-run dynamodb-local

build:
	CGO_ENABLED=0 go build -ldflags='-w -s' -o bin/$(SERVICE_NAME) .

run:
	go run .

tidy:
	go mod tidy

test:
	go test ./... -count=1

vet:
	go vet ./...

lint: vet

clean:
	rm -rf bin/

docker:
	docker build -t $(SERVICE_NAME):latest .

docker-run:
	docker run --rm -p $(PORT):$(PORT) --env-file app.env $(SERVICE_NAME):latest

dynamodb-local:
	docker run -d --name dynamodb-local -p 8000:8000 amazon/dynamodb-local:latest \
		-jar DynamoDBLocal.jar -sharedDb -inMemory
