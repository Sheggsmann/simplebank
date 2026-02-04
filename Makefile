postgres:
	docker run --name simplebank-postgres -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:alpine3.23

createdb:
	docker exec -it simplebank-postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simeplebank-posgres drobdb --username=root --owner=root simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5433/simple_bank?sslmode=disable" -verbose down

sqlc:
	sqlc generate

test:
	go test -v -cover ./...


.PHONY: postgres createdb migrateup migratedown