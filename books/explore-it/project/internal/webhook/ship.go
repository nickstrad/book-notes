package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"queuelet/internal/store"
	"sync"
	"time"
)

// Ship attempts up to 20 due rows, least recently scheduled first, concurrently.
// A pass costs at most one client deadline, and rows that keep failing move behind
// rows not yet tried, so slow destinations cannot hold the head of the outbox.
func Ship(s *store.Store, now time.Time) error {
	rows, e := s.Rows(`SELECT * FROM webhook_outbox WHERE delivered_at IS NULL AND next_at <= $1 ORDER BY next_at, id LIMIT 20`, now)
	if e != nil {
		return e
	}
	errs := make([]error, len(rows))
	var wg sync.WaitGroup
	for i, r := range rows {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = deliver(s, r, now)
		}()
	}
	wg.Wait()
	return errors.Join(errs...)
}
func deliver(s *store.Store, r map[string]any, now time.Time) error {
	body, _ := json.Marshal(r["body"])
	mac := hmac.New(sha256.New, []byte(r["secret"].(string)))
	mac.Write(body)
	req, e := http.NewRequest("POST", r["url"].(string), bytes.NewReader(body))

	var resp *http.Response
	if e == nil {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Queuelet-Signature", hex.EncodeToString(mac.Sum(nil)))
		resp, e = (&http.Client{Timeout: 2 * time.Second}).Do(req)
	}
	status := 0
	message := ""
	if resp != nil {
		status = resp.StatusCode
	}
	if e != nil {
		message = e.Error()
	} else if status < 200 || status >= 300 {
		message = fmt.Sprintf("HTTP %d", status)
	}
	if message != "" {
		if resp != nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		_, e = s.DB.Exec(`UPDATE webhook_outbox SET attempts=attempts+1,next_at=$2,event_at=$3,status=$4,error=$5 WHERE id=$1`, r["id"], now.Add(5*time.Second), now, status, message)
		return e
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	_, e = s.DB.Exec(`UPDATE webhook_outbox SET delivered_at=$2,event_at=$2,attempts=attempts+1,status=$3,error='' WHERE id=$1`, r["id"], now, status)
	return e
}
