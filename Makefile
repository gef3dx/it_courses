test:
	go test -v -race -count=1 ./...

build:
	go generate ./...
	go build ./...

run:
	go generate ./...
	go run ./cmd/api

migration:
	@test -n "$(name)" || (echo "usage: make migration name=create_users_table" && exit 1)
	@next=$$(find migrations -maxdepth 1 -type f -name '*.sql' | sed -E 's#.*/([0-9]{6})_.*#\1#' | sort | tail -n 1 | awk '{printf "%06d", $$1 + 1}'); \
	if [ -z "$$next" ]; then next=000001; fi; \
	file="migrations/$${next}_$(name).sql"; \
	printf -- "-- Write your migration here\n" > "$$file"; \
	echo "created $$file"
