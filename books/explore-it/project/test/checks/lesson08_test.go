//go:build checks

package checks

import (
	"os/exec"
	"strings"
	"testing"
)

// Each E08 check drives a neighbour of the whole lab (database, partner, clock), so the
// checks reuse the black-box reproductions and must not run in parallel.
func l08Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.Contains(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
func TestL08U36(t *testing.T) { l08Absent(t, "U36") }
func TestL08U37(t *testing.T) { l08Absent(t, "U37") }
func TestL08U38(t *testing.T) { l08Absent(t, "U38") }
func TestL08A5(t *testing.T)  { l08Absent(t, "A5") }
