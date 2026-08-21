make start:
	go run cmd/api/main.go

make seed:
	go run cmd/migrations/seed/main.go

pin-images:
	@echo "golang:1.25.13-alpine ->"
	@docker pull -q golang:1.25.13-alpine >/dev/null
	@docker inspect golang:1.25.13-alpine --format '{{index .RepoDigests 0}}'
	@echo "alpine:3.22 ->"
	@docker pull -q alpine:3.22 >/dev/null
	@docker inspect alpine:3.22 --format '{{index .RepoDigests 0}}'