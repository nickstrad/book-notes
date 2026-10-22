//go:build checks

package checks

// Shared public oracles also run unchanged against lesson-05 binaries.
import "testing"

func TestL06U27(t *testing.T) { l01Absent(t, "U27") }
func TestL06U28(t *testing.T) { l01Absent(t, "U28") }
func TestL06U29(t *testing.T) { l01Absent(t, "U29") }
func TestL06U30(t *testing.T) { l01Absent(t, "U30") }
func TestL06U31(t *testing.T) { l01Absent(t, "U31") }
func TestL06U32(t *testing.T) { l01Absent(t, "U32") }
func TestL06U33(t *testing.T) { l01Absent(t, "U33") }
func TestL06A4(t *testing.T)  { l01Absent(t, "A4") }
