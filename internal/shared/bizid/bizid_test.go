package bizid

import (
	"strconv"
	"strings"
	"testing"
)

func TestNewReturnsUUIDV7(t *testing.T) {
	id := New()

	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		t.Fatalf("New() = %q, want uuid string with 5 parts", id)
	}
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		t.Fatalf("New() = %q, want canonical uuid string", id)
	}
	if parts[2][0] != '7' {
		t.Fatalf("New() version = %q, want 7", parts[2][0])
	}
	variant, err := strconv.ParseUint(parts[3][:1], 16, 8)
	if err != nil {
		t.Fatalf("parse uuid variant: %v", err)
	}
	if variant < 8 || variant > 11 {
		t.Fatalf("New() variant nibble = %x, want RFC 4122 variant", variant)
	}
}
