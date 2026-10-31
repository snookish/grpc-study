# gRPC Playground

I built this repo to get comfortable with gRPC in Go without hiding the moving parts behind a framework. It is small on purpose. You can read the protos, see how the generated code is wired, start a server, and watch unary and streaming calls work.

## What I was practicing

I wanted one place to work through a few ideas:

- splitting protobuf definitions into versioned packages instead of one large file
- importing shared types across proto files
- generating Go code with a predictable layout
- writing service code that is easy to test without a live server
- using client streaming and server streaming in the same process
- keeping the server useful with health checks and reflection

## Layout

- `proto/common/v1/log.proto` holds the shared `LogLevel` enum
- `proto/person/v1/person.proto` holds the person messages and service
- `proto/streaming/v1/streaming.proto` holds the streaming service
- `gen/go` holds the generated Go packages
- `internal/person` holds the person service implementation
- `internal/streaming` holds the streaming service implementation
- `cmd` holds the server and client entry points

I keep generated code out of `internal` and `cmd` so the hand written code stays separate from anything protoc produces.

## How the protos fit together

Each proto package has a matching directory and a version suffix. That makes the package name, the directory, and the import path agree with each other. `common.v1` has no service. It exists only to hold the `LogLevel` enum so the streaming service can import it:

```proto
import "common/v1/log.proto";
import "google/protobuf/timestamp.proto";
```

The person service uses wrapper responses instead of returning `Person` directly. That is not just style. A single message used as the request or response for several calls makes the API harder to change later. With `SavePersonResponse` and `GetPersonResponse`, I can add fields to one call without touching the others.

The generated Go package names are aliased at the import site:

```go
personv1 "github.com/snookish/grpc/gen/go/person/v1"
```

That keeps call sites readable when several versioned packages appear in the same file.

## What the services do

The person service keeps records in a map guarded by a read write mutex. `SavePerson` trims the name, checks that the date of birth is present and valid, then stores the record. `GetPerson` returns a `NotFound` status when the name is missing. `ListPeople` returns everyone sorted by name.

The streaming service has two calls:

- `LogStream` accepts a stream of log entries and returns the number received
- `StreamServerTime` sends the current time on a fixed interval until the client goes away

Both calls validate their input. `LogStream` rejects a missing or invalid timestamp, an unspecified level, and an empty message. `StreamServerTime` rejects an interval that is zero or negative.

## Running it

Start the server:

```sh
go run ./cmd/server
```

Then run the person client in another shell:

```sh
go run ./cmd/client
```

The server also has a streaming client command in `cmd`. It defaults to a log stream demo and can switch to the server time stream. Both clients talk to the same server on port 50051 by default.

The server listens for `SIGINT` and `SIGTERM` and shuts down gracefully. It also registers the standard gRPC health service and server reflection, so tools like `grpcurl` can inspect it without extra setup.

## Generating code

The Makefile runs protoc with the proto directory as the import root and with the Go module path set. That is why generated files land under `gen/go` with the package paths the protos declare. After you edit a proto, run the generation step before you build. The same Makefile also has targets for formatting, vetting, testing, building, and running each command. Open it when you need the exact target names.

## Testing and checks

I keep the service logic in packages that do not depend on a running server, so most checks are fast:

```sh
go test ./...
go vet ./...
buf lint
```

The proto files also have a `buf.yaml` at the repo root. That file tells buf that `proto` is the module root, which keeps package names and directories aligned.

## Notes and tradeoffs

The store is in memory. It resets when the server restarts. I did that because the point of this repo is the RPC shape and the streaming behavior, not persistence.

I used `google.protobuf.Timestamp` for dates of birth. That is a little odd. A calendar date type would be a better fit, but the standard timestamp keeps the example focused on imports and generated types.

Both services share one port. In a larger system I would probably split them, give each one its own lifecycle, and add authentication and TLS. I left those out because this is local practice code.

The server uses a five second grace period during shutdown. That is enough for these demos, but a real service should tune that value against its slowest call.
