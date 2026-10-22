package metering

import (
	"database/sql"
	"encoding/json"
	"strings"

	"litefleet/internal/wire"
)

// Outbox is the queue of record: usage events the in-memory queue could not
// take, or could not ship, wait here in the node's own SQLite database until
// the sweeper delivers them. It survives restarts.
type Outbox struct {
	db *sql.DB
}

// NewOutbox opens the outbox table in db.
func NewOutbox(db *sql.DB) (*Outbox, error) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS usage_outbox (event_id text PRIMARY KEY, body text NOT NULL)`)
	return &Outbox{db: db}, err
}

// Add stores events; an event already there is left alone.
func (o *Outbox) Add(events []wire.UsageEvent) error {
	if len(events) == 0 {
		return nil
	}
	tx, err := o.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, ev := range events {
		body, _ := json.Marshal(ev)
		if _, err := tx.Exec(`INSERT OR IGNORE INTO usage_outbox (event_id, body) VALUES (?, ?)`, ev.EventID, string(body)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Take returns up to n stored events, oldest first.
func (o *Outbox) Take(n int) ([]wire.UsageEvent, error) {
	rows, err := o.db.Query(`SELECT body FROM usage_outbox ORDER BY rowid LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []wire.UsageEvent
	for rows.Next() {
		var body string
		var ev wire.UsageEvent
		if err := rows.Scan(&body); err != nil {
			return nil, err
		}
		if json.Unmarshal([]byte(body), &ev) == nil {
			out = append(out, ev)
		}
	}
	return out, rows.Err()
}

// Delete removes events that have been delivered.
func (o *Outbox) Delete(events []wire.UsageEvent) error {
	if len(events) == 0 {
		return nil
	}
	ids := make([]any, len(events))
	for i, ev := range events {
		ids[i] = ev.EventID
	}
	_, err := o.db.Exec(`DELETE FROM usage_outbox WHERE event_id IN (?`+strings.Repeat(",?", len(ids)-1)+`)`, ids...)
	return err
}

// Len is the number of events waiting in the outbox.
func (o *Outbox) Len() int {
	var n int
	o.db.QueryRow(`SELECT count(*) FROM usage_outbox`).Scan(&n)
	return n
}
