// Package a2a defines the explicitly implemented A2A version metadata contract.
// It does not translate cards or runtime wire messages.
package a2a

import (
	"encoding/json"
	"fmt"
)

func Supported(version string) bool { return version == "0.3" || version == "1.0" }

// CardVersion checks all protocol declarations, not method names or the mere
// presence of supportedInterfaces. Card schema validation remains the CP's job.
func CardVersion(raw string) (string, error) {
	var card map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &card) != nil || card == nil {
		return "", fmt.Errorf("agent card must be a JSON object")
	}
	legacy, hasLegacy := card["protocolVersion"]
	interfaces, hasInterfaces := card["supportedInterfaces"]
	if hasLegacy && hasInterfaces {
		return "", fmt.Errorf("contradictory card versions: protocolVersion and supportedInterfaces cannot coexist")
	}
	if hasLegacy {
		var version string
		if json.Unmarshal(legacy, &version) != nil || (version != "0.3" && version != "0.3.0") {
			return "", fmt.Errorf("unsupported legacy protocolVersion %s; supported legacy versions are 0.3 and 0.3.0", legacy)
		}
		return "0.3", nil
	}
	if hasInterfaces {
		var entries []struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if json.Unmarshal(interfaces, &entries) != nil || len(entries) == 0 {
			return "", fmt.Errorf("supportedInterfaces must declare at least one A2A 1.0 interface")
		}
		for _, entry := range entries {
			if entry.ProtocolVersion != "1.0" {
				return "", fmt.Errorf("unsupported supportedInterfaces protocolVersion %q; only 1.0 is implemented", entry.ProtocolVersion)
			}
		}
		return "1.0", nil
	}
	return "", fmt.Errorf("agent card has no protocol version: declare protocolVersion 0.3/0.3.0 or supportedInterfaces with protocolVersion 1.0")
}

// Resolve preserves the active-card schema version as a separate concern.
// Legacy records derive the runtime version solely from their original bytes;
// their empty advertised version defaults to the native 1.0 proxy frontend.
func Resolve(activeVersion, original, runtimeVersion, advertisedVersion string) (runtime, advertised string, err error) {
	if !Supported(activeVersion) {
		return "", "", fmt.Errorf("unsupported active-card a2aVersion %q; supported versions are 0.3 and 1.0", activeVersion)
	}
	if original == "" {
		return "", "", fmt.Errorf("originalCardRaw is missing; refresh the runtime Agent Card before selecting an adapter")
	}
	detected, err := CardVersion(original)
	if err != nil {
		return "", "", fmt.Errorf("originalCardRaw: %w", err)
	}
	if runtimeVersion == "" {
		runtimeVersion = detected
	}
	if !Supported(runtimeVersion) {
		return "", "", fmt.Errorf("unsupported runtimeA2AVersion %q; supported versions are 0.3 and 1.0", runtimeVersion)
	}
	if runtimeVersion != detected {
		return "", "", fmt.Errorf("runtimeA2AVersion %q contradicts originalCardRaw version %q; refresh the original runtime card and its version together", runtimeVersion, detected)
	}
	if advertisedVersion == "" {
		advertisedVersion = "1.0"
	}
	if !Supported(advertisedVersion) {
		return "", "", fmt.Errorf("unsupported advertisedA2AVersion %q; supported versions are 0.3 and 1.0", advertisedVersion)
	}
	if activeVersion == "1.0" && advertisedVersion == "0.3" {
		return "", "", fmt.Errorf("advertising a 1.0 active card as 0.3 is not implemented; provide a validated 0.3 active card or advertise 1.0")
	}
	return runtimeVersion, advertisedVersion, nil
}
