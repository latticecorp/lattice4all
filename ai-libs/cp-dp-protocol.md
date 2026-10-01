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
