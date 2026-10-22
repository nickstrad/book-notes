package httpx

import (
	"net/http"
	"sort"
	"strings"
)

// Health states, worst last. A service is degraded when it still serves but
// something it depends on is not right; failed when it cannot serve.
const (
	OK       = "ok"
	Degraded = "degraded"
	Failed   = "failed"
)

// Check reports one aspect of a service's health: a state and, unless ok,
// what is wrong.
type Check func() (state, detail string)

// HealthHandler runs every check and answers {status, checks}: 200 for ok and
// degraded (the instance should keep receiving traffic), 503 for failed (the
// balancer should take it out).
func HealthHandler(checks map[string]Check) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := OK
		results := map[string]string{}
		var problems []string
		names := make([]string, 0, len(checks))
		for name := range checks {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			state, detail := checks[name]()
			results[name] = state
			if state != OK {
				results[name] = state + ": " + detail
				problems = append(problems, detail)
			}
			if rank(state) > rank(status) {
				status = state
			}
		}
		body := map[string]any{"status": status, "checks": results}
		if len(problems) > 0 {
			body["detail"] = strings.Join(problems, "; ")
		}
		code := http.StatusOK
		if status == Failed {
			code = http.StatusServiceUnavailable
		}
		WriteJSON(w, code, body)
	}
}

func rank(state string) int {
	switch state {
	case Degraded:
		return 1
	case Failed:
		return 2
	}
	return 0
}
