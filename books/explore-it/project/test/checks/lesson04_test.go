//go:build checks

package checks

import (
	"os/exec"
	"strings"
	"testing"
)

func l04Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.HasPrefix(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
func TestL04U19(t *testing.T) { l04Absent(t, "U19") }
func TestL04U20(t *testing.T) { l04Absent(t, "U20") }
func TestL04U21(t *testing.T) { l04Absent(t, "U21") }
func TestL04U22(t *testing.T) { l04Absent(t, "U22") }
func TestL04U23(t *testing.T) { l04Absent(t, "U23") }
func TestL04A1(t *testing.T)  { l04Absent(t, "A1") }

// TestL04Invariants drives 200 random jobs and asserts the E04 rules after every round.
func TestL04Invariants(t *testing.T) {
	cmd := exec.Command("python3", "lab/repro/invariants04.py", "I04")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.HasPrefix(string(out), "INVARIANTS HOLD: ") {
		t.Fatalf("invariant walk: %v\n%s", e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
