module money-module

go 1.26.4

require (
	github.com/jackc/pgx/v5 v5.10.0
	google.golang.org/grpc v1.82.0
	google.golang.org/protobuf v1.36.11
	warchest/protos v0.0.0-00010101000000-000000000000
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260414002931-afd174a4e478 // indirect
)

replace warchest/protos => ../protos
