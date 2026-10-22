//go:build checks

package checks

import (
	"os/exec"
	"strings"
	"testing"
)

func l02Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, e := cmd.CombinedOutput()
	if e != nil || !strings.HasPrefix(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}
func TestL02T01(t *testing.T) { l02Absent(t, "T01") }
func TestL02T02(t *testing.T) { l02Absent(t, "T02") }
func TestL02T03(t *testing.T) { l02Absent(t, "T03") }
func TestL02U09(t *testing.T) { l02Absent(t, "U09") }
func TestL02U10(t *testing.T) { l02Absent(t, "U10") }
