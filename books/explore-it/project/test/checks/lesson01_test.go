//go:build checks

package checks

import (
	"os/exec"
	"strings"
	"testing"
)

// E01 regressions execute the same public behavior oracles as historical replays.
// These tests deliberately run serially: partner/outage fixtures affect the lab.
// Sharing an oracle prevents regression assertions and historical evidence drifting.
func l01Absent(t *testing.T, id string) {
	t.Helper()
	cmd := exec.Command("bash", "lab/repro/"+id+".sh")
	cmd.Dir = "../.."
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.HasPrefix(string(out), "ABSENT: "+id+" — ") {
		t.Fatalf("%s regression: %v\n%s", id, err, out)
	}
	t.Log(strings.TrimSpace(string(out)))
}

func TestL01D01(t *testing.T) { l01Absent(t, "D01") }
func TestL01D02(t *testing.T) { l01Absent(t, "D02") }
func TestL01D03(t *testing.T) { l01Absent(t, "D03") }
func TestL01D04(t *testing.T) { l01Absent(t, "D04") }
func TestL01D05(t *testing.T) { l01Absent(t, "D05") }
func TestL01D06(t *testing.T) { l01Absent(t, "D06") }
func TestL01D07(t *testing.T) { l01Absent(t, "D07") }
func TestL01D08(t *testing.T) { l01Absent(t, "D08") }
func TestL01D09(t *testing.T) { l01Absent(t, "D09") }
func TestL01D10(t *testing.T) { l01Absent(t, "D10") }
func TestL01D11(t *testing.T) { l01Absent(t, "D11") }
func TestL01D12(t *testing.T) { l01Absent(t, "D12") }
func TestL01D13(t *testing.T) { l01Absent(t, "D13") }
func TestL01D14(t *testing.T) { l01Absent(t, "D14") }
func TestL01D15(t *testing.T) { l01Absent(t, "D15") }
func TestL01D16(t *testing.T) { l01Absent(t, "D16") }
func TestL01D17(t *testing.T) { l01Absent(t, "D17") }
func TestL01D18(t *testing.T) { l01Absent(t, "D18") }
func TestL01D19(t *testing.T) { l01Absent(t, "D19") }
func TestL01D20(t *testing.T) { l01Absent(t, "D20") }
func TestL01D21(t *testing.T) { l01Absent(t, "D21") }
func TestL01D22(t *testing.T) { l01Absent(t, "D22") }
func TestL01D23(t *testing.T) { l01Absent(t, "D23") }
func TestL01D24(t *testing.T) { l01Absent(t, "D24") }
func TestL01D25(t *testing.T) { l01Absent(t, "D25") }
func TestL01D26(t *testing.T) { l01Absent(t, "D26") }
func TestL01D27(t *testing.T) { l01Absent(t, "D27") }
func TestL01U01(t *testing.T) { l01Absent(t, "U01") }
func TestL01U02(t *testing.T) { l01Absent(t, "U02") }
func TestL01U03(t *testing.T) { l01Absent(t, "U03") }
func TestL01U04(t *testing.T) { l01Absent(t, "U04") }
func TestL01U05(t *testing.T) { l01Absent(t, "U05") }
func TestL01U06(t *testing.T) { l01Absent(t, "U06") }
func TestL01U07(t *testing.T) { l01Absent(t, "U07") }
func TestL01U08(t *testing.T) { l01Absent(t, "U08") }
