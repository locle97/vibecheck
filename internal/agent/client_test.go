package agent

import (
	"strings"
	"testing"
)

func TestNew_UnknownBinary(t *testing.T) {
	_, err := New("unknown")
	if err == nil {
		t.Fatal("expected error for unknown binary")
	}
}

func TestNew_KnownBinaries(t *testing.T) {
	tests := []string{"claude", "cursor-agent", "opencode"}
	for _, bin := range tests {
		a, err := New(bin)
		if err != nil {
			t.Errorf("binary %q: unexpected error: %v", bin, err)
		}
		if a == nil {
			t.Errorf("binary %q: got nil agent", bin)
		}
	}
}

func TestUnwrapPrintJSON_EnvelopeResult(t *testing.T) {
	inner := `[{"id":1}]`
	raw := `{"type":"result","subtype":"success","is_error":false,"result":"` +
		strings.ReplaceAll(inner, `"`, `\"`) + `"}`

	got, err := unwrapPrintJSON(raw, "claude")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != inner {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestUnwrapPrintJSON_EnvelopeError(t *testing.T) {
	raw := `{"type":"result","subtype":"error","is_error":true,"result":"bad key"}`
	_, err := unwrapPrintJSON(raw, "claude")
	if err == nil {
		t.Fatal("expected error when is_error=true")
	}
}

func TestUnwrapPrintJSON_FallbackRaw(t *testing.T) {
	raw := `[{"id":1}]`
	got, err := unwrapPrintJSON(raw, "claude")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != raw {
		t.Fatalf("unexpected result: %q", got)
	}
}

func TestExtractNDJSONContent(t *testing.T) {
	inner := `[{"id":1}]`
	raw := `{"type":"thinking","content":"Analyzing diff..."}` + "\n" +
		`{"type":"text","content":"` + strings.ReplaceAll(inner, `"`, `\"`) + `"}` + "\n" +
		`{"type":"done"}`

	got := extractNDJSONContent(raw)
	if got != inner {
		t.Fatalf("unexpected content: %q", got)
	}
}
