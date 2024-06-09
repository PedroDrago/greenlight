PORT=-port=4000
ENV=-env=dev
FLAGS=$(ENV) $(PORT) $(LIMITER)
all: air

air:
	air
run:
	go run ./cmd/api $(FLAGS)
air:
	cd cmd/api && air
