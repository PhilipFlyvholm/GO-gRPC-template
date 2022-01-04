# GO-gRPC-template

## Proto command:

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative routeguide/route.proto
```

## Start server:

```
go run ./server/server.go 0
go run ./server/server.go 1
go run ./server/server.go 2
```

## Frontend commands:

```
go run ./frontend/frontend.go 0
go run ./frontend/frontend.go 1
```

## Client commands:

```
go run ./client/client.go
```