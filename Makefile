make start:
	go run cmd/api/main.go

make seed:
	go run cmd/migrations/seed/main.go