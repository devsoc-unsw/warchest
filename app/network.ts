import * as path from 'node:path';
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';

const PROTO_PATH = path.join(__dirname, 'helloworld.proto');

const packageDefinition = protoLoader.loadSync(PROTO_PATH, {
  keepCase: true,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true,
});

// We cast to 'any' here for dynamic loading. In production, 
// you'd typically generate static types for this object.
const helloProto = grpc.loadPackageDefinition(packageDefinition).helloworld as any;

// Define the interfaces based on your .proto file
interface HelloRequest {
  name: string;
}

interface HelloReply {
  message: string;
}

/**
 * Implements the SayHello RPC method with strict typing.
 */
function sayHello(
  call: grpc.ServerUnaryCall<HelloRequest, HelloReply>,
  callback: grpc.sendUnaryData<HelloReply>
) {
  callback(null, { message: `Hello ${call.request.name}!` });
}

function main() {
  const server = new grpc.Server();
  server.addService(helloProto.Greeter.service, { sayHello });
  
  const port = '0.0.0.0:50051';
  server.bindAsync(port, grpc.ServerCredentials.createInsecure(), (err, portNumber) => {
    if (err) {
      console.error(err);
      return;
    }
    console.log(`Server running at http://0.0.0.0:${portNumber}`);
  });
}

main();