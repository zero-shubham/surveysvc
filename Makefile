get_migrate:
	go get -u github.com/golang-migrate/migrate/cmd/migrate

create_migrate: 
	migrate create -ext sql -dir db/migrations/ -seq $(seq)

migrate_up:
	migrate -path db/migrations/ -database "postgresql://postgres:postgres@localhost:5432/datab?sslmode=disable" -verbose up

migrate_down:
	migrate -path db/migrations/ -database "postgresql://postgres:postgres@localhost:5432/datab?sslmode=disable" -verbose down

seed:
	go run ./cmd/seed/main.go