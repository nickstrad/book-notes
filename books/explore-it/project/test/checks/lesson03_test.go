//go:build checks

package checks

import (
	"os/exec"
	"strings"
	"testing"
)

func l03Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.HasPrefix(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
func TestL03U11(t *testing.T) { l03Absent(t, "U11") }
func TestL03U12(t *testing.T) { l03Absent(t, "U12") }
func TestL03U13(t *testing.T) { l03Absent(t, "U13") }
func TestL03U14(t *testing.T) { l03Absent(t, "U14") }
func TestL03U15(t *testing.T) { l03Absent(t, "U15") }
func TestL03U16(t *testing.T) { l03Absent(t, "U16") }
func TestL03U17(t *testing.T) { l03Absent(t, "U17") }
func TestL03U18(t *testing.T) { l03Absent(t, "U18") }
