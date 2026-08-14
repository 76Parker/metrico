package signature

import "testing"

func TestSum(t *testing.T) {
	got := Sum([]byte(`[{"id":"Alloc"}]`), "test-key")
	const expected = "57ea53ead42bc53489f9e7e572a6d7064d8e3d28b8a1dcc7acfc197eb017eb34"

	if got != expected {
		t.Fatalf("Sum() = %q, want %q", got, expected)
	}
}
