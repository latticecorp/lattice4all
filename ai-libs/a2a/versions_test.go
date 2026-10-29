package a2a

import "testing"

func TestCardVersionsAreExplicit(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`{"protocolVersion":"0.3.0"}`, "0.3"},
		{`{"protocolVersion":"0.3"}`, "0.3"},
		{`{"supportedInterfaces":[{"protocolVersion":"1.0"}]}`, "1.0"},
		{`{"protocolVersion":"0.3.1"}`, ""},
		{`{"supportedInterfaces":[{"protocolVersion":"1.1"}]}`, ""},
		{`{"supportedInterfaces":[{"protocolVersion":"1.0"},{"protocolVersion":"2.0"}]}`, ""},
		{`{"protocolVersion":"0.3","supportedInterfaces":[{"protocolVersion":"1.0"}]}`, ""},
		{`{"supportedInterfaces":[]}`, ""},
		{`{"supportedInterfaces":[{}]}`, ""},
		{`{}`, ""},
	} {
		actual, err := CardVersion(tc.raw)
		if actual != tc.want || ((err == nil) != (tc.want != "")) {
			t.Fatalf("%s: %q %v", tc.raw, actual, err)
		}
	}
}

func TestRuntimeAndAdvertisedVersionsAreIndependent(t *testing.T) {
	original := `{"protocolVersion":"0.3.0"}`
	runtime, advertised, err := Resolve("1.0", original, "0.3", "1.0")
	if err != nil || runtime != "0.3" || advertised != "1.0" {
		t.Fatalf("%s %s %v", runtime, advertised, err)
	}
	runtime, advertised, err = Resolve("0.3", original, "", "")
	if err != nil || runtime != "0.3" || advertised != "1.0" {
		t.Fatalf("legacy: %s %s %v", runtime, advertised, err)
	}
	for _, versions := range [][2]string{{"1.0", "1.0"}, {"0.4", "1.0"}, {"0.3", "1.1"}} {
		if _, _, err := Resolve("1.0", original, versions[0], versions[1]); err == nil {
			t.Fatalf("contradictory versions accepted: %v", versions)
		}
	}
	if _, _, err := Resolve("1.0", "", "", ""); err == nil {
		t.Fatal("runtime inferred without original")
	}
	if _, _, err := Resolve("1.0", original, "0.3", "0.3"); err == nil {
		t.Fatal("unimplemented down-projection accepted")
	}
}
