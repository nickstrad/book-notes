package jobs

import (
	"fmt"
	"slices"
	"strings"
)

// The job lifecycle as one table. Every statement that changes a job's state
// (or removes the job) takes its state guard from here, so a missing guard is
// a missing table row rather than a forgotten WHERE term. Lease-holder checks
// (token, expiry) and the choice between several targets stay with the caller.
const (
	Scheduled = "scheduled"
	Queued    = "queued"
	Leased    = "leased"
	RetryWait = "retry_wait"
	Succeeded = "succeeded"
	Dead      = "dead"
	Cancelled = "cancelled"
	// Gone is the target of events that delete the row.
	Gone = "gone"
)

var States = []string{Scheduled, Queued, Leased, RetryWait, Succeeded, Dead, Cancelled}

type Event string

const (
	EvPromote   Event = "promote"   // run_at reached
	EvLease     Event = "lease"     // a worker leases the job
	EvHeartbeat Event = "heartbeat" // the holder extends its lease
	EvAck       Event = "ack"       // the holder succeeds
	EvNack      Event = "nack"      // the holder fails: retry_wait, or dead when spent, refused or purged
	EvExpire    Event = "expire"    // the lease ran out: queued, or dead when spent or purged
	EvRelease   Event = "release"   // the holder was deregistered: as expire
	EvCancel    Event = "cancel"    // operator cancel
	EvRetry     Event = "retry"     // operator retry of a dead letter
	EvPurge     Event = "purge"     // pending rows are deleted; leased rows are marked and stay leased
	EvDelete    Event = "delete"    // operator delete of a finished job
	EvRetain    Event = "retain"    // retention sweep
)

type arrow struct{ from, to []string }

var table = map[Event][]arrow{
	EvPromote:   {{[]string{Scheduled, RetryWait}, []string{Queued}}},
	EvLease:     {{[]string{Queued}, []string{Leased}}},
	EvHeartbeat: {{[]string{Leased}, []string{Leased}}},
	EvAck:       {{[]string{Leased}, []string{Succeeded}}},
	EvNack:      {{[]string{Leased}, []string{RetryWait, Dead}}},
	EvExpire:    {{[]string{Leased}, []string{Queued, Dead}}},
	EvRelease:   {{[]string{Leased}, []string{Queued, Dead}}},
	EvCancel:    {{[]string{Scheduled, Queued, Leased, RetryWait}, []string{Cancelled}}},
	EvRetry:     {{[]string{Dead}, []string{Queued}}},
	EvPurge: {
		{[]string{Scheduled, Queued, RetryWait}, []string{Gone}},
		{[]string{Leased}, []string{Leased}},
	},
	EvDelete: {{[]string{Succeeded, Dead, Cancelled}, []string{Gone}}},
	// Dead letters are exempt from automatic retention (A4).
	EvRetain: {{[]string{Succeeded, Cancelled}, []string{Gone}}},
}

// Events lists every event in the table, in a stable order.
func Events() []Event {
	out := make([]Event, 0, len(table))
	for ev := range table {
		out = append(out, ev)
	}
	slices.Sort(out)
	return out
}

// Targets returns the states ev may move a job in state from to; nil means the
// event is refused in that state.
func Targets(ev Event, from string) []string {
	for _, a := range table[ev] {
		if slices.Contains(a.from, from) {
			return a.to
		}
	}
	return nil
}

// From returns every state in which ev is accepted, optionally only those it
// moves to target.
func From(ev Event, target ...string) []string {
	out := []string{}
	for _, a := range table[ev] {
		if len(target) > 0 && !slices.ContainsFunc(target, func(t string) bool { return slices.Contains(a.to, t) }) {
			continue
		}
		out = append(out, a.from...)
	}
	if len(out) == 0 {
		panic(fmt.Sprintf("jobs: event %s has no transition to %v", ev, target))
	}
	return out
}

// Guard is the SQL state guard for ev on column col, e.g. "j.state IN ('leased')".
// The states are package constants, never input, so inlining them is safe.
func Guard(col string, ev Event, target ...string) string {
	return col + " IN ('" + strings.Join(From(ev, target...), "','") + "')"
}
