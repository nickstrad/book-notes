package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
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
		if e != nil || resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if resp != nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			_, e = s.DB.Exec(`UPDATE webhook_outbox SET attempts=attempts+1,next_at=$2 WHERE id=$1`, r["id"], now.Add(5*time.Second))
			if e != nil {
				return e
			}
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		log.Printf("webhook delivered status=%d", resp.StatusCode)
		_, e = s.DB.Exec(`UPDATE webhook_outbox SET delivered_at=$2,attempts=attempts+1 WHERE id=$1`, r["id"], now)
		if e != nil {
			return e
		}
	}
	return nil
}
