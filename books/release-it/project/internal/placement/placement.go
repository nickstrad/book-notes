// Package placement decides which node a new database lives on.
package placement

import (
	"errors"
	"sort"
)

// ErrNoNodes means no node has sent a recent heartbeat.
var ErrNoNodes = errors.New("no live nodes")

// Pick returns the live node holding the fewest databases (ties by name).
func Pick(load map[string]int) (string, error) {
	if len(load) == 0 {
		return "", ErrNoNodes
	}
	names := make([]string, 0, len(load))
	for n := range load {
		names = append(names, n)
	}
	sort.Strings(names)
	best := names[0]
	for _, n := range names[1:] {
		if load[n] < load[best] {
			best = n
		}
	}
	return best, nil
}
