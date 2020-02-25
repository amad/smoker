package version

import (
	"testing"
)

func TestString(t *testing.T) {
	if String() != version {
		t.Fatalf("Version.String() invalid format\nexpected: %s\nreceived: %s", version, String())
	}
}
