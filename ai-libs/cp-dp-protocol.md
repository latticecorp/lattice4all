# Control Plane - Dataplane Protocol

This document details the bidirectional gRPC streaming protocol (`ControlPlaneService.SubscribeDataplaneConfiguration`) used between the Control Plane and Dataplane instances (e.g., `inference-router`, `ai-agent`).

The protocol is designed to deliver linked pairs of configurations (xDS for Envoy and Dataplane configuration for the app layer) and collect detailed delivery, validation, and operational feedback.

## Dataplane Messages

The Dataplane communicates with the Control Plane by sending a stream of `DataplaneMessage` objects. There are three types of payloads the Dataplane sends: `Subscribe`, `UpdateStatus`, and `EnvoyStatus`.

### 1. `Subscribe`
**Purpose:** Initiates the stream, authenticates the instance, and establishes the checkpoint for state resumption.

*   **Identity & Authorization:** Sends `service_id` and `service_instance_id`. The Control Plane uses these to identify the logical service and specific instance, authorizing them against the connection's identity (e.g., mTLS certificates).
*   **Dataplane Type:** Sends `kind` (e.g., `DATAPLANE_KIND_INFERENCE_ROUTER`) so the Control Plane knows which schema/features this dataplane expects.
*   **State Resumption:** Sends `last_dataplane_version`, which represents the last generation where *both* xDS and Dataplane configurations were successfully stored locally.
    *   If this matches the Control Plane's current target version, the Control Plane can skip a full state-of-the-world replay.
    *   If left empty (or if internal versions differ), it forces a full replay from the Control Plane.

### 2. `UpdateStatus`
**Purpose:** Provides immediate, synchronous feedback on whether a configuration update from the Control Plane was successfully parsed, validated, and stored in the Dataplane's memory.

*   **Targeted Feedback:** The `target` field specifies whether the feedback corresponds to an `xds` or `dataplane` configuration payload.
*   **Memory Checkpoint:** It confirms receipt and semantic validation (`accepted = true`). It does *not* mean Envoy has applied the configuration yet.
*   **Diagnostic Errors:** If the update is malformed or violates business rules (e.g., duplicate route tokens), it sets `accepted = false` and provides an `error` string. This error string is strictly bounded and sanitized to ensure it never leaks raw configurations or user prompts.

### 3. `EnvoyStatus`
**Purpose:** Forwards the asynchronous xDS ACK/NACK feedback from the local Envoy proxy back to the Control Plane.

*   **Operational Readiness:** Envoy applies its configurations (CDS, EDS, LDS, RDS) asynchronously. When Envoy accepts or rejects a specific configuration, the Dataplane captures this event and bridges it to the Control Plane via `EnvoyStatus`.
*   **Resource Granularity:** It includes the `type_url` (the specific xDS resource type), along with the `version`, `nonce`, and whether it was `accepted`.
*   **Convergence:** The Control Plane relies on this message to determine when the proxy layer has actually converged and is ready to route traffic based on the new generation.

## A2A card and adapter version metadata

The accepted policy is A2A 1.0 native with explicit 0.3 compatibility. The additive
fields in the catalog and `RegistryAgentCard` separate:

- `a2a_version`: the schema of the stored active card (unchanged).
- `runtime_a2a_version`: the runtime codec version, checked against original bytes.
- `advertised_a2a_version`: the proxy frontend/discovery version.
- `original_card_raw` in `RegistryAgentCard`: exact original runtime bytes, distinct
  from `active_card_raw`, which remains authoritative for governed discovery.

`WellKnownMetadata` carries runtime/advertised versions at ingestion and refresh.
Internal forwarded refresh requests preserve those captured values. The CP validates
both cards and contradictory version declarations; refresh updates original bytes
and runtime version together without replacing curation or advertised metadata.
Only canonical `0.3` and `1.0` metadata are implemented. Legacy card wire version
`0.3.0` normalizes to `0.3`. Future interface versions are rejected explicitly.

For legacy records, absent runtime metadata is resolved solely from a validated
`original_card_raw`; missing advertised metadata defaults to `1.0`. Missing or
ambiguous originals are errors, never permission to use the curated/caller version.
Old protobuf readers ignore these additive fields. New proxies accept completely
legacy bundles using their existing strict catalog read, but check all fields when
new bundle metadata is present. Stale metadata must not overwrite a newer catalog
snapshot. Bundle sizes now include two raw card payloads per card.

Proxies own runtime wire translation and supported card projection. The CP has no
A2A execution RPC surface. Proxy callers use `A2A-Version: 1.0` for native requests;
missing/empty headers select 0.3 independently of method names and runtime adapters.
Unsupported projection features are rejected, not silently dropped.

Rollout note: additive protobuf fields do not imply mixed-binary Raft command or
snapshot compatibility. Older Control Plane binaries strictly decode persisted
JSON and do not recognize the new domain fields. Quiesce card ingestion/refresh
and curation while upgrading CP voters; do not introduce versioned card writes
until all voters understand the fields. Back up data before upgrade. Downgrading
after new metadata has been persisted requires restoring a compatible pre-upgrade
backup; no destructive snapshot rewrite is performed by this change.
