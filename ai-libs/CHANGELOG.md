# Changelog

## v0.1.0-dev.2

- Breaking: rename the streaming RPC to `SubscribeDataplaneConfiguration`.
- Replace `router_id` with required `service_id` and `service_instance_id`.
- Merge resume fields into `last_dataplane_version`, the last jointly stored
  xDS/dataplane generation; an empty checkpoint requests a full replay.
- Rename `RouterMessage` to `DataplaneMessage`, `RouterConfiguration` to
  `DataplaneConfiguration`, and response field `router_configuration` to
  `dataplane_configuration`. Feedback target is now `dataplane`.
- Regenerate Go bindings. CatalogService remains unchanged.
- Both streaming clients and servers must upgrade together. dev.1 remains
  immutable for Go proxy/checksum compatibility.

## v0.1.0-dev.1

- Extract the public CatalogService schema from `lattice-cp` commit
  `f0d6097c171894fd2660ac65a8e843c93f03610c` with all eight RPCs unchanged.
- Centralize the existing development ControlPlaneService streaming contract
  from `intelligent-inference-router` without changing its wire interface.
- Publish generated Go protobuf and gRPC client/server bindings as
  `github.com/latticecorp/lattice4all/ai-libs`.
- Pin generation tools and document versioned dependencies for both consumers.

This release does not add a ControlPlaneService implementation to `lattice-cp`.
Its CatalogService is available to the router through the shared client bindings.
