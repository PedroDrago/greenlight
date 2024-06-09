PORT=-port=4000
ENV=-env=dev
FLAGS=$(ENV) $(PORT) $(LIMITER)
all: 
	go run ./cmd/api/

run:
	go run ./cmd/api $(FLAGS)
air:
	cd cmd/api && air

migrate:
	migrate -path=./migrations -database=$GREENLIGHT_DB up
