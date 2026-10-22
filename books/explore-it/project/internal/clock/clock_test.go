package clock

import (
	"testing"
	"time"
)

func TestSemanticClock(t *testing.T) {
	var c Clock
	at := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	if _, e := c.Change("set", at, 0); e != nil {
		t.Fatal(e)
	}
	if !c.Now().Equal(at) {
		t.Fatal("set did not freeze")
	}
	if _, e := c.Change("advance", time.Time{}, 3600); e != nil {
		t.Fatal(e)
	}
	if !c.Now().Equal(at.Add(time.Hour)) {
		t.Fatal("wrong advance")
	}
	if _, e := c.Change("advance", time.Time{}, 1e20); e == nil {
		t.Fatal("overflow accepted")
	}
	if _, e := c.Change("reset", time.Time{}, 0); e != nil {
		t.Fatal(e)
	}
	if time.Since(c.Now()) > time.Second {
		t.Fatal("reset did not restore real time")
	}
}
