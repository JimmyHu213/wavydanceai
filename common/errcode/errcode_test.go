package errcode

import (
	"regexp"
	"testing"
)

func TestCodesAreUniqueAndWellFormed(t *testing.T) {
	codes := All()
	if len(codes) == 0 {
		t.Fatal("All() returned no codes")
	}
	format := regexp.MustCompile(`^[a-z]+(\.[a-z_]+)+$`)
	seen := map[string]bool{}
	for _, c := range codes {
		if !format.MatchString(c) {
			t.Errorf("code %q is not dotted-lowercase (e.g. domain.reason)", c)
		}
		if seen[c] {
			t.Errorf("duplicate code value %q", c)
		}
		seen[c] = true
	}
}
