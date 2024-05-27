PORT=-port=4000
ENV=-env=dev
FLAGS=$(ENV) $(PORT)
run:
	go run ./cmd/api $(FLAGS)
air:
	cd cmd/api && air
