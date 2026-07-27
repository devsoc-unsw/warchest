# How to use this directory:
Protos are the 'protocols' which are used to send data between different services.
Think of each as a custom 'route' which is used to send data from a to b.

To generate a proto. First define the schema in `your_service.proto`.

Then run the following command. (If you don't have protoc just spin up the dev container)
```Bash
mkdir your_service &&
protoc --go_out=your_service --go_opt=paths=source_relative \
       --go-grpc_out=your_service --go-grpc_opt=paths=source_relative \
       your_service.proto
```

This will generate two such files:
`your_service.pb.go`: contains all of the data structures
`your_service_grpc.pb.go`: contains the interface that needs to be implemented

To migrate:
Modify the `your_service.proto`, then remove the `your_service` dir

i.e
```Bash
rm -rf your_service
```

then regenerate the protos (same as above)
```Bash
mkdir your_service &&
protoc --go_out=your_service --go_opt=paths=source_relative \
       --go-grpc_out=your_service --go-grpc_opt=paths=source_relative \
       your_service.proto
```