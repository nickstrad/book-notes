//go:build checks

package checks

import (
	"os/exec"
	"strings"
	"testing"
)

func l05Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.HasPrefix(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
func TestL05U24(t *testing.T) { l05Absent(t, "U24") }
func TestL05U25(t *testing.T) { l05Absent(t, "U25") }
func TestL05U26(t *testing.T) { l05Absent(t, "U26") }
func TestL05A2(t *testing.T)  { l05Absent(t, "A2") }

// TestL05Walk takes 100 random verbs in random orders and asserts the E05 rules after every step.
func TestL05Walk(t *testing.T) {
	cmd := exec.Command("python3", "lab/repro/walk05.py", "W05")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.HasPrefix(string(out), "WALK HOLDS: 100/100 steps") {
		t.Fatalf("random walk: %v\n%s", e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
