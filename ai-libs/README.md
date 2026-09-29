# Lattice AI libraries

This directory is the source of truth for the public protobuf contracts and
generated Go clients/server interfaces used by `lattice-cp` and
`intelligent-inference-router`. Change schemas here and regenerate bindings here;
consumers must not maintain copies of these schemas or generated files.

## Use from Go

Requires Go 1.26 or newer. Generated bindings are committed; consumers do not
need protoc.

```sh
go get github.com/latticecorp/lattice4all/ai-libs@v0.1.0-dev.2
```

```go
import (
    catalogv1 "github.com/latticecorp/lattice4all/ai-libs/api/lattice/catalog/v1"
    controlv1 "github.com/latticecorp/lattice4all/ai-libs/api/lattice/control/v1"
)
```

| Package | Contract | Implementation |
| --- | --- | --- |
| `api/lattice/catalog/v1` | `lattice.catalog.v1.CatalogService`: create, get, list, update, delete, transactions, operation lookup, health | `lattice-cp` |
| `api/lattice/control/v1` | `lattice.control.v1.ControlPlaneService`: bidirectional router configuration delivery and feedback | Router development server; not yet implemented by `lattice-cp` |

Use `catalogv1.NewCatalogServiceClient(conn)` to call the control plane, and
`catalogv1.RegisterCatalogServiceServer(server, implementation)` to serve it.
Catalog calls require `authorization: Bearer TOKEN` and `x-lattice-scope`
outgoing gRPC metadata; mutations additionally require `x-request-id`.
Use a deadline and TLS credentials in deployed clients. The development control
plane listens on `127.0.0.1:9400` with plaintext explicitly enabled.
Configure maximum received message size to 16 MiB to match catalog responses.
Follow `next_cursor` to retrieve all list pages; strict reads use `stale=false`.

The router's `internal/controlplane.CatalogClient` and `cmd/catalog-discover`
provide authenticated, paginated catalog discovery using this module. Catalog
records alone do not define xDS resources, route generations, backend ports, or
model routing policy; the separate streaming configuration API remains necessary
for that existing router functionality.

## Migrate consumers

1. Add the versioned module dependency with the command above.
2. Replace `github.com/latticecorp/lattice-cp/api/lattice/catalog/v1` imports with
   the shared catalog package. Replace
   `github.com/latticecorp/intelligent-inference-router/api/lattice/control/v1`
   imports with the shared control package.
3. Remove the corresponding local `.proto` and `.pb.go` files and local generation
   rules. Run `go mod tidy` and the consumer's tests.

CatalogService is unchanged. The development streaming API has breaking changes
in dev.2; all stream clients and servers must upgrade together. See below. Do not import old and new
generated packages into the same binary: they register the same protobuf names.

For local development across the three sibling checkouts, create an uncommitted
Go workspace from their parent directory:

```sh
go work init ./lattice4all/ai-libs ./lattice-cp ./intelligent-inference-router
```

Committed consumer modules should depend on a release without local `replace`
directives. Publish this module before resolving that version in consumers.

## Generate and verify

Install protoc **36.2**, then run from this directory:

```sh
make tools  # protoc-gen-go v1.36.12; protoc-gen-go-grpc v1.6.2
make proto
make check
```

Commit schema and generated binding changes together. Keep changes within v1
wire compatible; never reuse removed field numbers or names.

## Development releases

The Go module is in a subdirectory, so tags must include `ai-libs/`:
`ai-libs/v0.1.0-dev.2`. The version used by `go get` excludes that prefix.
Development releases are GitHub prereleases. Increment the dev suffix for each
release; never move an existing tag. See [release notes](CHANGELOG.md).

## Dataplane subscription (dev.2)

The shared `lattice.control.v1.ControlPlaneService` RPC is now
`SubscribeDataplaneConfiguration(stream DataplaneMessage) returns (stream ControlPlaneMessage)`.
The first message contains `Subscribe` with required `service_id` (logical
catalog service), `service_instance_id` (running instance), `schema_version: 1`,
and one `last_dataplane_version`. Servers must authorize both identity fields
against the authenticated caller; the development server only validates them.

`RouterMessage` and `RouterConfiguration` are now `DataplaneMessage` and
`DataplaneConfiguration`; the response field is `dataplane_configuration` and
update feedback uses target `"dataplane"` (or `"xds"`). Old RPCs/types are removed.

xDS and dataplane configuration for a generation share one immutable version.
The resume checkpoint is sent only when both configurations are stored with
that version; otherwise it is empty and the server must replay both complete
configurations. Stage dataplane configuration before publishing xDS. Delivery
acceptance remains separate from Envoy ACK/NACK and readiness. A matching
checkpoint must still receive content replays or another supported lease renewal
before freshness expires; the development server replays both payloads every 20s.
