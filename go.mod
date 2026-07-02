module saga

go 1.18

require (
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	google.golang.org/grpc v1.62.1
	google.golang.org/protobuf v1.33.0
)

// Добавляем директиву replace для проблемного пакета
replace google.golang.org/genproto => github.com/googleapis/go-genproto v0.0.0-20230822172742-b8732ec3820d

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
)
