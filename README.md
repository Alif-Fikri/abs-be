create migration : goose create create_tes1_table sql

command goose : running, status, drop

goose -dir database/migrations mysql "root:password@tcp(127.0.0.1:3306)/sekolahdb" up

goose -dir database/migrations mysql "root:password@tcp(127.0.0.1:3306)/sekolahdb" status

goose -dir database/migrations mysql "root:password@tcp(127.0.0.1:3306)/sekolahdb" down

go run main.go
