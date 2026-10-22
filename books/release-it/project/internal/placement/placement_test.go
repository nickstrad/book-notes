package placement

import "testing"

func TestPickFewest(t *testing.T) {
	got, err := Pick(map[string]int{"node-a": 5, "node-b": 3})
	if err != nil || got != "node-b" {
		t.Fatalf("Pick = %q, %v; want node-b", got, err)
	}
}

func TestPickTieByName(t *testing.T) {
	got, _ := Pick(map[string]int{"node-b": 2, "node-a": 2})
	if got != "node-a" {
		t.Fatalf("Pick = %q; want node-a", got)
	}
}

func TestPickNoNodes(t *testing.T) {
	if _, err := Pick(nil); err != ErrNoNodes {
		t.Fatalf("err = %v; want ErrNoNodes", err)
	}
}
