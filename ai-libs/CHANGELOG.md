# Changelog

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
