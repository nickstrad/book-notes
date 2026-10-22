package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"queuelet/internal/store"
	"time"
)

func Ship(s *store.Store, now time.Time) error {
	rows, e := s.Rows(`SELECT * FROM webhook_outbox WHERE delivered_at IS NULL AND next_at <= $1 ORDER BY id LIMIT 20`, now)
	if e != nil {
		return e
	}
	for _, r := range rows {
		body, _ := json.Marshal(r["body"])
		mac := hmac.New(sha256.New, []byte(r["secret"].(string)))
		mac.Write(body)
		req, e := http.NewRequest("POST", r["url"].(string), bytes.NewReader(body))
		if e != nil {
			return e
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Queuelet-Signature", hex.EncodeToString(mac.Sum(nil)))
		resp, e := (&http.Client{Timeout: 2 * time.Second}).Do(req)
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
			if e != nil {
				return e
			}
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		_, e = s.DB.Exec(`UPDATE webhook_outbox SET delivered_at=$2,event_at=$2,attempts=attempts+1,status=$3,error='' WHERE id=$1`, r["id"], now, status)
		if e != nil {
			return e
		}
	}
	return nil
}
