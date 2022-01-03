# GO-gRPC-template

## Proto command:

```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative routeguide/route.proto
```

## Start server:

```
go run ./server/server.go
```

## Client server:

```
go run ./client/client.go
```
